package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	cetypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	sdk "github.com/engineerdna/engineerdna/plugins/plugin-sdk"
)

type AWSCostsPlugin struct {
	accessKeyID     string
	secretAccessKey string
	region          string
	costTags        []string
	awsConfig       aws.Config
}

func (p *AWSCostsPlugin) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		Name:        "aws-costs",
		Version:     "0.1.0",
		Type:        "source",
		Description: "AWS infrastructure costs and resource metrics",
		Author:      "EngineerDNA",
		ConfigFields: []sdk.ConfigField{
			{
				Name:        "aws_access_key_id",
				Type:        "password",
				Required:    true,
				Description: "AWS Access Key ID",
				Secret:      true,
			},
			{
				Name:        "aws_secret_access_key",
				Type:        "password",
				Required:    true,
				Description: "AWS Secret Access Key",
				Secret:      true,
			},
			{
				Name:        "region",
				Type:        "string",
				Required:    false,
				Description: "AWS Region",
				Default:     "us-east-1",
			},
			{
				Name:        "cost_allocation_tags",
				Type:        "string",
				Required:    false,
				Description: "Comma-separated cost allocation tags (e.g., team,project,env)",
			},
		},
		Anonymization: sdk.AnonymizationSpec{
			Required: false,
			Reason:   "No PII in cost data",
			Strategy: "sequential",
			Fields:   []string{},
		},
	}
}

func (p *AWSCostsPlugin) Configure(cfg map[string]string) error {
	accessKeyID, ok := cfg["aws_access_key_id"]
	if !ok || accessKeyID == "" {
		return fmt.Errorf("aws_access_key_id is required")
	}
	p.accessKeyID = accessKeyID

	secretAccessKey, ok := cfg["aws_secret_access_key"]
	if !ok || secretAccessKey == "" {
		return fmt.Errorf("aws_secret_access_key is required")
	}
	p.secretAccessKey = secretAccessKey

	// Region is optional, default to us-east-1
	p.region = cfg["region"]
	if p.region == "" {
		p.region = "us-east-1"
	}

	// Parse comma-separated cost allocation tags
	if tags, ok := cfg["cost_allocation_tags"]; ok && tags != "" {
		parts := strings.Split(tags, ",")
		var validTags []string
		for _, tag := range parts {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				validTags = append(validTags, trimmed)
			}
		}
		p.costTags = validTags
	}

	// Create AWS config with static credentials
	ctx := context.Background()
	var err error
	p.awsConfig, err = config.LoadDefaultConfig(ctx,
		config.WithRegion(p.region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			p.accessKeyID,
			p.secretAccessKey,
			"", // session token (empty for static credentials)
		)),
	)
	if err != nil {
		return fmt.Errorf("failed to create AWS config: %w", err)
	}

	return nil
}

func (p *AWSCostsPlugin) Health() sdk.HealthResult {
	if p.accessKeyID == "" || p.secretAccessKey == "" {
		return sdk.HealthResult{
			Healthy: false,
			Message: "Not configured",
		}
	}

	// Quick health check: verify credentials with STS GetCallerIdentity (via EC2 DescribeRegions)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ec2Client := ec2.NewFromConfig(p.awsConfig)
	_, err := ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return sdk.HealthResult{
			Healthy: false,
			Message: fmt.Sprintf("AWS API error: %v", err),
		}
	}

	return sdk.HealthResult{
		Healthy: true,
		Message: fmt.Sprintf("Connected to AWS (region: %s)", p.region),
	}
}

func (p *AWSCostsPlugin) Sync(params sdk.SyncParams) (sdk.SyncResult, error) {
	if p.accessKeyID == "" {
		return sdk.SyncResult{}, fmt.Errorf("plugin not configured")
	}

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var allMetrics []sdk.Metric
	var warnings []string

	// Calculate time range for sync
	sinceTime := params.Since
	if sinceTime.IsZero() {
		// First sync: fetch last 30 days
		sinceTime = time.Now().UTC().AddDate(0, 0, -30)
	}

	// 1. Fetch costs from Cost Explorer
	costMetrics, err := p.fetchCosts(ctx, sinceTime)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Cost fetch failed: %v", err))
		fmt.Fprintf(os.Stderr, "Cost fetch error: %v\n", err)
	} else {
		allMetrics = append(allMetrics, costMetrics...)
	}

	// 2. Fetch EC2 instance counts (current snapshot)
	ec2Metrics, err := p.fetchEC2Counts(ctx)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("EC2 count failed: %v", err))
		fmt.Fprintf(os.Stderr, "EC2 count error: %v\n", err)
	} else {
		allMetrics = append(allMetrics, ec2Metrics...)
	}

	// 3. Fetch RDS instance counts
	rdsMetrics, err := p.fetchRDSCounts(ctx)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("RDS count failed: %v", err))
		fmt.Fprintf(os.Stderr, "RDS count error: %v\n", err)
	} else {
		allMetrics = append(allMetrics, rdsMetrics...)
	}

	// 4. Fetch S3 storage metrics
	s3Metrics, err := p.fetchS3Storage(ctx, sinceTime)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("S3 storage failed: %v", err))
		fmt.Fprintf(os.Stderr, "S3 storage error: %v\n", err)
	} else {
		allMetrics = append(allMetrics, s3Metrics...)
	}

	// 5. Fetch Lambda invocation counts
	lambdaMetrics, err := p.fetchLambdaInvocations(ctx, sinceTime)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Lambda invocations failed: %v", err))
		fmt.Fprintf(os.Stderr, "Lambda invocations error: %v\n", err)
	} else {
		allMetrics = append(allMetrics, lambdaMetrics...)
	}

	return sdk.SyncResult{
		Metrics:  allMetrics,
		Warnings: warnings,
	}, nil
}

func (p *AWSCostsPlugin) fetchCosts(ctx context.Context, since time.Time) ([]sdk.Metric, error) {
	ceClient := costexplorer.NewFromConfig(p.awsConfig)

	// Format dates for Cost Explorer (YYYY-MM-DD)
	start := since.Format("2006-01-02")
	end := time.Now().UTC().Format("2006-01-02")

	// Request cost data grouped by service
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &cetypes.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: cetypes.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []cetypes.GroupDefinition{
			{
				Type: cetypes.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
		},
	}

	result, err := ceClient.GetCostAndUsage(ctx, input)
	if err != nil {
		return nil, sdk.NewAuthError(fmt.Sprintf("Cost Explorer API error: %v", err))
	}

	var metrics []sdk.Metric
	for _, resultByTime := range result.ResultsByTime {
		// Parse timestamp from YYYY-MM-DD format
		timestamp, err := time.Parse("2006-01-02", *resultByTime.TimePeriod.Start)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse timestamp %s: %v\n", *resultByTime.TimePeriod.Start, err)
			continue
		}

		for _, group := range resultByTime.Groups {
			// Extract service name
			service := "unknown"
			if len(group.Keys) > 0 {
				service = group.Keys[0]
			}

			// Extract cost value
			costStr := ""
			if group.Metrics != nil {
				if unblendedCost, ok := group.Metrics["UnblendedCost"]; ok {
					costStr = *unblendedCost.Amount
				}
			}

			// Parse cost as float
			var costValue float64
			if _, err := fmt.Sscanf(costStr, "%f", &costValue); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse cost %s: %v\n", costStr, err)
				continue
			}

			// Skip zero costs
			if costValue == 0 {
				continue
			}

			metric := sdk.Metric{
				MetricName:  "aws_cost_total",
				Timestamp:   timestamp.UTC(),
				Granularity: "daily",
				Value:       costValue,
				Unit:        "dollars",
				Dimensions: map[string]interface{}{
					"service": service,
					"region":  p.region,
				},
			}
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

func (p *AWSCostsPlugin) fetchEC2Counts(ctx context.Context) ([]sdk.Metric, error) {
	ec2Client := ec2.NewFromConfig(p.awsConfig)

	// Describe running instances
	input := &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running"},
			},
		},
	}

	result, err := ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, sdk.NewNetworkError(fmt.Sprintf("EC2 API error: %v", err))
	}

	// Count instances by type
	instanceCounts := make(map[string]int)
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instanceType := string(instance.InstanceType)
			instanceCounts[instanceType]++
		}
	}

	// Create metrics
	var metrics []sdk.Metric
	now := time.Now().UTC()
	for instanceType, count := range instanceCounts {
		metric := sdk.Metric{
			MetricName:  "ec2_instance_count",
			Timestamp:   now,
			Granularity: "daily",
			Value:       float64(count),
			Unit:        "count",
			Dimensions: map[string]interface{}{
				"region":        p.region,
				"instance_type": instanceType,
			},
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (p *AWSCostsPlugin) fetchRDSCounts(ctx context.Context) ([]sdk.Metric, error) {
	rdsClient := rds.NewFromConfig(p.awsConfig)

	// Describe DB instances
	result, err := rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, sdk.NewNetworkError(fmt.Sprintf("RDS API error: %v", err))
	}

	// Count instances by engine
	engineCounts := make(map[string]int)
	for _, instance := range result.DBInstances {
		engine := "unknown"
		if instance.Engine != nil {
			engine = *instance.Engine
		}
		engineCounts[engine]++
	}

	// Create metrics
	var metrics []sdk.Metric
	now := time.Now().UTC()
	for engine, count := range engineCounts {
		metric := sdk.Metric{
			MetricName:  "rds_instance_count",
			Timestamp:   now,
			Granularity: "daily",
			Value:       float64(count),
			Unit:        "count",
			Dimensions: map[string]interface{}{
				"region": p.region,
				"engine": engine,
			},
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (p *AWSCostsPlugin) fetchS3Storage(ctx context.Context, since time.Time) ([]sdk.Metric, error) {
	cwClient := cloudwatch.NewFromConfig(p.awsConfig)

	// Query CloudWatch for S3 storage metrics
	// Note: S3 metrics are in CloudWatch under AWS/S3 namespace
	// BucketSizeBytes metric with StandardStorage storage type
	now := time.Now().UTC()
	startTime := since

	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/S3"),
		MetricName: aws.String("BucketSizeBytes"),
		Dimensions: []types.Dimension{
			{
				Name:  aws.String("StorageType"),
				Value: aws.String("StandardStorage"),
			},
		},
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(now),
		Period:     aws.Int32(86400), // 1 day in seconds
		Statistics: []types.Statistic{types.StatisticAverage},
	}

	result, err := cwClient.GetMetricStatistics(ctx, input)
	if err != nil {
		return nil, sdk.NewNetworkError(fmt.Sprintf("CloudWatch API error: %v", err))
	}

	// Convert datapoints to metrics
	var metrics []sdk.Metric
	for _, datapoint := range result.Datapoints {
		if datapoint.Average == nil || datapoint.Timestamp == nil {
			continue
		}

		// Convert bytes to gigabytes
		sizeGB := *datapoint.Average / (1024 * 1024 * 1024)

		metric := sdk.Metric{
			MetricName:  "s3_storage_gb",
			Timestamp:   datapoint.Timestamp.UTC(),
			Granularity: "daily",
			Value:       sizeGB,
			Unit:        "gigabytes",
			Dimensions: map[string]interface{}{
				"region": p.region,
			},
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (p *AWSCostsPlugin) fetchLambdaInvocations(ctx context.Context, since time.Time) ([]sdk.Metric, error) {
	lambdaClient := lambda.NewFromConfig(p.awsConfig)
	cwClient := cloudwatch.NewFromConfig(p.awsConfig)

	// List Lambda functions
	listResult, err := lambdaClient.ListFunctions(ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		return nil, sdk.NewNetworkError(fmt.Sprintf("Lambda API error: %v", err))
	}

	var metrics []sdk.Metric
	now := time.Now().UTC()

	// For each function, get invocation metrics from CloudWatch
	for _, function := range listResult.Functions {
		if function.FunctionName == nil {
			continue
		}

		funcName := *function.FunctionName

		// Query CloudWatch for invocation count
		input := &cloudwatch.GetMetricStatisticsInput{
			Namespace:  aws.String("AWS/Lambda"),
			MetricName: aws.String("Invocations"),
			Dimensions: []types.Dimension{
				{
					Name:  aws.String("FunctionName"),
					Value: aws.String(funcName),
				},
			},
			StartTime:  aws.Time(since),
			EndTime:    aws.Time(now),
			Period:     aws.Int32(86400), // 1 day in seconds
			Statistics: []types.Statistic{types.StatisticSum},
		}

		result, err := cwClient.GetMetricStatistics(ctx, input)
		if err != nil {
			// Log error but continue with other functions
			fmt.Fprintf(os.Stderr, "Failed to fetch metrics for function %s: %v\n", funcName, err)
			continue
		}

		// Convert datapoints to metrics
		for _, datapoint := range result.Datapoints {
			if datapoint.Sum == nil || datapoint.Timestamp == nil {
				continue
			}

			metric := sdk.Metric{
				MetricName:  "lambda_invocations",
				Timestamp:   datapoint.Timestamp.UTC(),
				Granularity: "daily",
				Value:       *datapoint.Sum,
				Unit:        "count",
				Dimensions: map[string]interface{}{
					"region":        p.region,
					"function_name": funcName,
				},
			}
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

func main() {
	sdk.Serve(&AWSCostsPlugin{})
}
