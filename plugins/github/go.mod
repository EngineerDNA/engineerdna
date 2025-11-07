module github.com/engineerdna/engineerdna/plugins/github

go 1.25.3

replace github.com/engineerdna/engineerdna/plugins/plugin-sdk => ../plugin-sdk

require (
	github.com/engineerdna/engineerdna/plugins/plugin-sdk v0.0.0-00010101000000-000000000000
	github.com/google/go-github/v57 v57.0.0
	github.com/google/uuid v1.6.0
	golang.org/x/oauth2 v0.32.0
)

require github.com/google/go-querystring v1.1.0 // indirect
