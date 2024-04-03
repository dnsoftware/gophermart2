package app

import (
	"context"
	"fmt"
	"github.com/dnsoftware/gophermart2/internal/gophermart/config"
	"github.com/dnsoftware/gophermart2/internal/gophermart/domain"
	"github.com/dnsoftware/gophermart2/internal/gophermart/handlers"
	"github.com/dnsoftware/gophermart2/internal/logger"
	"github.com/dnsoftware/gophermart2/internal/storage"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func Run() {
	var wg sync.WaitGroup
	gLogger := logger.Log()
	defer gLogger.Sync()

	cfg := config.NewServerConfig()

	// репозитории
	martStorage, err := storage.NewMartStorage(cfg.DatabaseURI)
	if err != nil {
		panic(err)
	}

	userRepo := storage.NewUserRepo(martStorage)
	orderRepo := storage.NewOrderRepo(martStorage)
	balanceRepo := storage.NewBalanceRepo(martStorage)
	accrualRepo := storage.NewAccrualRepo(cfg.AccrualAddress)

	// канал с ордерами на проверку
	chanUnchecked := domain.NewOrdersUnchecked()
	// канал с проверенными ордерами
	chanChecked := domain.NewOrdersChecked()

	// основные объекты
	user := domain.NewUserModel(userRepo)

	// берет из chanBalance и сохраняет в базу
	balance := domain.NewBalanceModel(balanceRepo)

	// готовит ордера на проверку chanUnchecked<-, берет данные проверенных ордеров и сохраняет статус в базу ордеров <-chanChecked
	// а также в базу балансов
	order := domain.NewOrderModel(orderRepo, chanUnchecked, chanChecked, balance)

	// отсылает ордера на проверку <-chanUnchecked, ставит в очередь на сохранение chanChecked<-
	accrual := domain.NewAccrualModel(accrualRepo, chanUnchecked, chanChecked)

	ctxSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	order.ProcessUnchecked(ctxSignal)
	order.ProcessChecked(ctxSignal)

	// работа с accrual сервисом - проверка начислений по заказам
	wg.Add(1)
	go func() {
		defer wg.Done()
		accrual.StartAccrualChecker(ctxSignal)
	}()

	srv := handlers.NewServer(cfg.RunAddress, user, order, balance)

	// запуск HTTP сервера
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Listening and serving")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// ждет сигнала завершения программы, потом завершает работу HTTP сервера и работу с Accrual
	wg.Add(1)
	go func() {
		<-ctxSignal.Done()

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer func() {
			wg.Done()
			stop()
			cancel()
		}()

		if err := srv.Shutdown(ctxTimeout); err != nil {
			logger.Log().Error("Ошибка при завершении HTTP сервера: " + err.Error())
			return
		}

		fmt.Println("Shutdown completed")
	}()

	wg.Wait()

	fmt.Println("\nПрограмма завершена!")

}
