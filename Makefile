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
	@staticcheck -go 1.22 ./...

check_deps:
	go install mvdan.cc/gofumpt@v0.6.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.1
	go install github.com/daixiang0/gci@v0.12.1
	go install honnef.co/go/tools/cmd/staticcheck@v0.4.6

test:
	go test ./...

update:
	go get -u ./...
