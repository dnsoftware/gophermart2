package domain

import (
	"context"
	"github.com/dnsoftware/gophermart2/internal/constants"
)

/* Очередь ордеров на проверку */

type OrdersUnchecked struct {
	ordersCh chan int64
}

func NewOrdersUnchecked() *OrdersUnchecked {
	return &OrdersUnchecked{
		ordersCh: make(chan int64, constants.OrdersChannelCapacity),
	}
}

// ставим в очередь для дальнейшей отправки на проверку
func (u *OrdersUnchecked) Push(number int64) {
	u.ordersCh <- number
}

// забираем из очереди для отпарвки на проверку в Accrual
func (u *OrdersUnchecked) Pop(ctx context.Context) int64 {
	select {
	case <-ctx.Done():
	case o := <-u.ordersCh:
		return o
	}

	return 0
}
