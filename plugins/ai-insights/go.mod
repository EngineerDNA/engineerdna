module github.com/engineerdna/engineerdna/plugins/ai-insights

go 1.25.3

replace github.com/engineerdna/engineerdna/plugins/plugin-sdk => ../plugin-sdk

replace github.com/engineerdna/engineerdna => ../..

require (
	github.com/engineerdna/engineerdna v0.0.0-00010101000000-000000000000
	github.com/engineerdna/engineerdna/plugins/plugin-sdk v0.0.0-00010101000000-000000000000
)
