package domain

import (
	"context"
	"fmt"
	"github.com/dnsoftware/gophermart2/internal/constants"
	"github.com/dnsoftware/gophermart2/internal/logger"
	"github.com/dnsoftware/gophermart2/internal/storage"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type AccrualStorage interface {
	GetOrder(orderNum int64) (*storage.AccrualRow, int, error)
}

// непроверенные ордера берем и отсылаем на проверку
type Unchecked interface {
	Pop(ctx context.Context) int64
}

// проверенные ордера ставим в очередь на сохранение
type Checked interface {
	Push(order int64, status string, accrual float32)
}

type AccrualItem struct {
}

type Accrual struct {
	storage                  AccrualStorage
	accrualServiceQueryLimit int           // максимально кол-во запросов к Accrual сервису в минуту
	counter                  int           // счетчик запросов за текущий интервал
	checkPeriod              time.Duration // период проверки
	ordersToCheck            Unchecked     // отсюда забираем ордера на проверку и шлем в Accrual
	ordersToSave             Checked       // сюда заносим проверенные ордера, полученные из Accrual
}

func NewAccrualModel(storage AccrualStorage, ordersToCheck Unchecked, ordersToSave Checked) *Accrual {
	balance := &Accrual{
		storage:                  storage,
		accrualServiceQueryLimit: constants.AccrualServiceQueryLimit,
		counter:                  0,
		checkPeriod:              time.Duration(constants.AccrualCheckPeriod) * time.Second,
		ordersToCheck:            ordersToCheck,
		ordersToSave:             ordersToSave,
	}

	return balance
}

// StartAccrualChecker Служба проверки начислений
func (b *Accrual) StartAccrualChecker(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(constants.AccrualTickerPeriod) * time.Second)

	for {
		select {
		case <-ctx.Done():
			logger.Log().Info("AccrualChecker DONE!!!")
			return
		case <-ticker.C:
			b.counter = 0
		default:
			if b.counter >= b.accrualServiceQueryLimit {
				continue
			}

			// основная работа
			orderNumber := b.ordersToCheck.Pop(ctx)
			order, status, err := b.storage.GetOrder(orderNumber)

			switch status {
			case http.StatusOK:

				switch order.Status {
				case constants.AccrualRegistered, constants.AccrualProcessing: // еще не готовы
					b.ordersToSave.Push(orderNumber, constants.OrderProcessing, 0)
				case constants.AccrualProcessed:
					b.ordersToSave.Push(orderNumber, constants.OrderProcessed, order.Accrual)
				case constants.AccrualInvalid:
					b.ordersToSave.Push(orderNumber, constants.OrderInvalid, order.Accrual)
				}

			case http.StatusNoContent:
				logger.Log().Info(fmt.Sprintf("Accrual GetOrder no content: %v", orderNumber))

			case http.StatusTooManyRequests:
				re := regexp.MustCompile(`^No more than (\d+) requests per minute allowed, Retry-After: (\d+)$`)
				matches := re.FindStringSubmatch(err.Error())
				if len(matches) < 3 {
					logger.Log().Error("Error regexp match accrual too many requests")
					break
				}

				newQueryLimit, _ := strconv.Atoi(matches[1])
				b.accrualServiceQueryLimit = newQueryLimit

				newCheckPeriod, _ := strconv.Atoi(matches[2])
				b.checkPeriod = time.Duration(newCheckPeriod) * time.Second
				ticker.Reset(b.checkPeriod)
				b.counter = b.accrualServiceQueryLimit // в этом временном отрезке запросов уже не будет

				logger.Log().Info(fmt.Sprintf("too many requests: %v", orderNumber))

			case http.StatusInternalServerError:
				logger.Log().Error("Accrual GetOrder error: " + err.Error())
			}

			b.counter++
		}
	}

}
