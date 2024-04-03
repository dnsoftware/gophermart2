package domain

import (
	"context"
	"fmt"
	"github.com/dnsoftware/gophermart2/internal/constants"
	"github.com/dnsoftware/gophermart2/internal/logger"
	"github.com/dnsoftware/gophermart2/internal/storage"
	"strconv"
	"time"
)

type OrderStorage interface {
	Create(ctx context.Context, userID, number int64) (int, error)
	List(ctx context.Context, userID int64) ([]storage.OrderRow, int, error)
	GetUnchecked(ctx context.Context) ([]storage.OrderRow, error)
	UpdateStatus(ctx context.Context, orderNumber int64, orderStatus string) error
	GetOrderByNumber(ctx context.Context, orderNumber int64) (storage.OrderRow, error)
}

// для работы с каналом непроверенных ордеров
type UncheckedOrders interface {
	Push(number int64)
}

type CheckedOrders interface {
	Pop(ctx context.Context) (int64, string, float32)
}

type BalanceAdd interface {
	AddTransaction(ctx context.Context, orderNumber int64, amount float32) error
}

type Order struct {
	storage       OrderStorage
	ordersToCheck UncheckedOrders // сюда кидаем номера ордеров на проверку в Accrual
	ordersToSave  CheckedOrders   // отсюда берем проверенные и сохраняем в базу
	balanceAdd    BalanceAdd      // для внесения начислений из проверенных ордеров на баланс
}

// OrderItem plain structure
type OrderItem struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

func NewOrderModel(storage OrderStorage, uncheckedCh UncheckedOrders, checkedCh CheckedOrders, balanceAdd BalanceAdd) *Order {
	order := &Order{
		storage:       storage,
		ordersToCheck: uncheckedCh,
		ordersToSave:  checkedCh,
		balanceAdd:    balanceAdd,
	}

	return order
}

func (o *Order) AddOrder(ctx context.Context, userID, number int64) (int, error) {
	status := constants.OrderInternalError

	// проверка Луна
	if !IsLuhnValid(number) {
		status = constants.OrderBadNumberFormat
		return status, fmt.Errorf("неверный номер заказа")
	}

	// сохраняем в базу
	status, err := o.storage.Create(ctx, userID, number)
	if err != nil {
		return status, err
	}

	logger.Log().Info(fmt.Sprintf("Добавлен заказ %v", number))

	return status, nil
}

func (o *Order) SetStatus(ctx context.Context, orderNumber int64, orderStatus string) error {

	err := o.storage.UpdateStatus(ctx, orderNumber, orderStatus)
	if err != nil {
		return fmt.Errorf("Ошибка при смене статуса заказа: " + err.Error())
	}

	return nil
}

func (o *Order) OrdersList(ctx context.Context, userID int64) ([]OrderItem, int, error) {
	orders := make([]OrderItem, 0)

	// получаем список заказов
	list, status, err := o.storage.List(ctx, userID)
	if err != nil {
		return []OrderItem{}, status, err
	}

	for _, item := range list {
		upAt := item.UploadedAt.Format(time.RFC3339)
		orderItem := OrderItem{
			Number:     strconv.FormatInt(item.Num, 10),
			Status:     item.Status,
			Accrual:    item.Accrual,
			UploadedAt: upAt,
		}

		orders = append(orders, orderItem)
	}

	return orders, status, nil
}

// постановка необработанных ордеров в очередь на проверку Accrual
func (o *Order) ProcessUnchecked(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Log().Info("ProcessUnchecked DONE!!!")
				return
			default:
				orders, err := o.storage.GetUnchecked(ctx)
				if err != nil {
					logger.Log().Error(err.Error())
				}

				for _, order := range orders {
					o.ordersToCheck.Push(order.Num)
				}
			}
		}
	}()
}

// Получение обработанных и сохранение в базу
func (o *Order) ProcessChecked(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Log().Info("ProcessChecked DONE!!!")
				return
			default:
				orderID, orderStatus, orderAccrual := o.ordersToSave.Pop(ctx)

				switch orderStatus {
				case constants.OrderInvalid, constants.OrderProcessing: // просто меняем статуc
					err := o.storage.UpdateStatus(ctx, orderID, orderStatus)
					if err != nil {
						logger.Log().Error(err.Error())
					}

				case constants.OrderProcessed: // меняем статус и отправляем в базу балансов
					// сохраняем в балансы
					err := o.balanceAdd.AddTransaction(ctx, orderID, orderAccrual)
					if err != nil {
						logger.Log().Error(err.Error())
					}
				}

			}
		}
	}()
}
