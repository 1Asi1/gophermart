run:
	go run ./cmd/gophermart/main.go -d postgres://asicloud:@localhost:5432/practicum?sslmode=disable

run_accrual:
	./cmd/accrual/accrual_darwin_arm64 -a :8081 -d postgres://asicloud:@localhost:5432/practicum

stop_accrual:
	pkill accrual_darwin_arm64

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -w /app \
    golangci/golangci-lint:v1.56.2-alpine \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json
