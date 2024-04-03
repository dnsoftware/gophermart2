package domain

import (
	"context"
	"fmt"
	"github.com/dnsoftware/gophermart2/internal/constants"
	"github.com/dnsoftware/gophermart2/internal/storage"
	"math"
	"strconv"
	"time"
)

type BalanceStorage interface {
	SaveTransaction(ctx context.Context, orderNumber int64, amount float32) error
	GetUserBalance(ctx context.Context, userID int64) (float32, error)
	GetUserWithdrawn(ctx context.Context, userID int64) (float32, error)
	GetUserWithdrawList(ctx context.Context, userID int64) ([]storage.WithdrawRow, error)
	WithdrawTransaction(ctx context.Context, userID int64, orderNumber int64, amount float32) error
}

type Balance struct {
	storage BalanceStorage
}

type CurrentBalance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type WithdrawItem struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func NewBalanceModel(storage BalanceStorage) *Balance {
	balance := &Balance{
		storage: storage,
	}

	return balance
}

// комплексное обновление данных в базе
func (b *Balance) AddTransaction(ctx context.Context, orderNumber int64, amount float32) error {

	err := b.storage.SaveTransaction(ctx, orderNumber, amount)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении начислений в базу %w", err)
	}

	return nil
}

func (b *Balance) UserBalance(ctx context.Context, userID int64) (*CurrentBalance, error) {
	balanceAmount, err := b.storage.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении баланса пользователя %w", err)
	}

	withdrawnAmount, err := b.storage.GetUserWithdrawn(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении использованных баллов пользователя %w", err)
	}

	return &CurrentBalance{
		Current:   balanceAmount,
		Withdrawn: withdrawnAmount,
	}, nil
}

func (b *Balance) UserWithrawalsList(ctx context.Context, userID int64) ([]WithdrawItem, error) {
	withdrawals, err := b.storage.GetUserWithdrawList(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка выводов пользователя %w", err)
	}

	var items []WithdrawItem
	for _, val := range withdrawals {
		item := WithdrawItem{
			Order:       strconv.FormatInt(val.Order, 10),
			Sum:         float32(math.Abs(float64(val.Sum))),
			ProcessedAt: val.ProcessedAt.Format(time.RFC3339),
		}

		items = append(items, item)
	}

	return items, nil
}

func (b *Balance) Withraw(ctx context.Context, userID int64, number int64, amount float32) (int, error) {
	// проверка на отрицательное списание
	if amount <= 0 {
		return constants.WithdrawNotEnoughFunds, fmt.Errorf("симма списания не может быть нулевой или отрицательной")
	}

	// проверка на корректность номера заказа
	if !IsLuhnValid(number) {
		return constants.WithdrawBadOrderNumber, fmt.Errorf("неверный номер заказа")
	}

	// проверка на достаточность средств
	balance, err := b.UserBalance(ctx, userID)
	if err != nil {
		return constants.WithdrawInternalError, fmt.Errorf("ошибка получения баланса: " + err.Error())
	}
	if amount > balance.Current {
		return constants.WithdrawNotEnoughFunds, fmt.Errorf(fmt.Sprintf("ошибка получения баланса, запрошено %v, в наличии %v: ", amount, balance.Current))
	}

	// обработка списания
	err = b.storage.WithdrawTransaction(ctx, userID, number, amount)
	if err != nil {
		return constants.WithdrawInternalError, fmt.Errorf("ошибка обработки списания: " + err.Error())
	}

	return constants.WithdrawalsOk, nil
}
