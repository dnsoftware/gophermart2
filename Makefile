PROJECT="gophermart"

default:
	echo ${PROJECT}

test:
	go test -v -count=1 ./...

.PHONY: default test cover
cover:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	rm coverage.out

.PHONY: gen
gen:
	mockgen -source=internal/gophermart/domain/user.go -destination=internal/gophermart/domain/mocks/mock_user_storage.go
	mockgen -source=internal/gophermart/domain/order.go -destination=internal/gophermart/domain/mocks/mock_order_storage.go
	mockgen -source=internal/gophermart/domain/accrual.go -destination=internal/gophermart/domain/mocks/mock_accrual_storage.go
	mockgen -source=internal/gophermart/domain/balance.go -destination=internal/gophermart/domain/mocks/mock_balance_storage.go