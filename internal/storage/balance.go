package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dnsoftware/gophermart2/internal/constants"
	"math"
	"time"
)

type BalanceRepo struct {
	storage *MartStorage
}

type BalanceRow struct {
	ID          int64
	UserID      int64
	OrderNumber int64
	Amount      float32
	ProcessedAt time.Time
}

type WithdrawRow struct {
	Order       int64
	Sum         float32
	ProcessedAt time.Time
}

func NewBalanceRepo(storage *MartStorage) *BalanceRepo {

	balance := BalanceRepo{
		storage: storage,
	}

	return &balance
}

func (b *BalanceRepo) SaveTransaction(ctx context.Context, orderNumber int64, amount float32) error {

	// получение ID владельца заказа
	var userID int64
	query := `SELECT user_id 
			  FROM orders WHERE num = $1`
	row := b.storage.db.QueryRowContext(ctx, query, orderNumber)

	err := row.Scan(&userID)
	if err != nil {
		return err
	}

	// проверка, что данные по этому заказу еще не вносились в базу
	var a float32
	query = `SELECT amount 
			 FROM balances WHERE order_number = $1`
	row = b.storage.db.QueryRowContext(ctx, query, orderNumber)

	err = row.Scan(&a)
	if err != nil && err != sql.ErrNoRows { // sql.ErrNoRows если нет строки
		return err
	}
	if err == nil { // если ошибок нет - то такая запись уже есть
		return fmt.Errorf("данные по начислению уже сохранены")
	}

	// стартуем транзакцию БД
	tx, err := b.storage.db.Begin()
	if err != nil {
		return err
	}

	// обновление статуса ордера на обработанный
	query = `UPDATE orders SET status = $1, accrual = $2
			  WHERE num = $3`

	err = b.storage.retryExec(ctx, query, constants.OrderProcessed, amount, orderNumber)
	if err != nil {
		tx.Rollback()
		return err
	}

	// занесение начислений на баланс
	query = `INSERT INTO balances (user_id, order_number, amount, processed_at)
			  VALUES ($1, $2, $3, now())`
	err = b.storage.retryExec(ctx, query, userID, orderNumber, amount)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (b *BalanceRepo) GetUserBalance(ctx context.Context, userID int64) (float32, error) {
	query := `SELECT SUM(amount) curr_balance 
			  FROM balances WHERE user_id = $1`
	row := b.storage.db.QueryRowContext(ctx, query, userID)

	var currBalance sql.NullFloat64
	err := row.Scan(&currBalance)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return float32(currBalance.Float64), nil
}

func (b *BalanceRepo) GetUserWithdrawn(ctx context.Context, userID int64) (float32, error) {
	query := `SELECT SUM(amount) curr_balance 
			  FROM balances WHERE user_id = $1 AND amount < 0`
	row := b.storage.db.QueryRowContext(ctx, query, userID)

	var withdrawBalance sql.NullFloat64
	err := row.Scan(&withdrawBalance)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return float32(math.Abs(withdrawBalance.Float64)), nil
}

func (b *BalanceRepo) GetUserWithdrawList(ctx context.Context, userID int64) ([]WithdrawRow, error) {
	query := `SELECT order_number, amount, processed_at 
			  FROM balances WHERE user_id = $1 AND amount < 0
			  ORDER BY processed_at ASC`
	rows, err := b.storage.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserWithdrawList error: %w", err)
	}
	defer rows.Close()

	var wd []WithdrawRow
	for rows.Next() {
		var w WithdrawRow
		err = rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("GetUserWithdrawList get row: %w", err)
		}

		wd = append(wd, w)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetUserWithdrawList rows.Err: %w", err)
	}

	return wd, nil
}

func (b *BalanceRepo) WithdrawTransaction(ctx context.Context, userID int64, orderNumber int64, amount float32) error {

	query := `INSERT INTO balances (user_id, order_number, amount, processed_at)
			  VALUES ($1, $2, $3, now())`
	err := b.storage.retryExec(ctx, query, userID, orderNumber, -amount)
	if err != nil {
		return err
	}

	return nil
}
