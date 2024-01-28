run:
	go run ./cmd/gophermart/main.go -d postgres://asicloud:@localhost:5432/practicum?sslmode=disable

run_accrual:
	./cmd/accrual/accrual_darwin_arm64 -a :8081 -d postgres://asicloud:@localhost:5432/practicum

stop_accrual:
	pkill accrual_darwin_arm64