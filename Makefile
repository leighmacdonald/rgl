all: fmt check

fmt:
	gci write . --skip-generated -s default
	gofumpt -l -w .

check: lint_golangci  static

lint_golangci:
	@golangci-lint run --timeout 3m ./...

fix: fmt
	@golangci-lint run --fix

static:
	@staticcheck -go 1.20 ./...

check_deps:
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3
	go install github.com/daixiang0/gci@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

test:
	go test ./...

update:
	go get -u ./...
