.PHONY: test
test:
	golangci-lint run
	go clean -testcache
	go test ./... -race -cover
	gosec ./...
	govulncheck ./...
