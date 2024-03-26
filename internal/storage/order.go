package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dnsoftware/gophermart2/internal/constants"
	"github.com/dnsoftware/gophermart2/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

type OrderRepo struct {
	storage *MartStorage
}

type OrderRow struct {
	ID         int64
	UserID     int64
	Num        int64
	Status     string
	Accrual    float32
	UploadedAt time.Time
}

func NewOrderRepo(storage *MartStorage) *OrderRepo {

	repo := OrderRepo{
		storage: storage,
	}

	return &repo
}

// Загрузка номера заказа
// возвращает стутус операции и ошибку
func (p *OrderRepo) Create(ctx context.Context, userID int64, number int64) (int, error) {

	query := `SELECT id, user_id FROM orders WHERE num = $1`
	row := p.storage.db.QueryRowContext(ctx, query, number)

	var id, uid int64

	err := row.Scan(&id, &uid)
	if err != nil && err != sql.ErrNoRows {
		return constants.OrderInternalError, err
	}

	if id > 0 && uid == userID { // этот пользователь уже добавил этот заказ, возвращаем OK
		return constants.OrderOk, nil
	}

	if id > 0 && uid != userID { // другой пользователь уже добавил этот заказ
		return constants.OrderAlreadyUpload, fmt.Errorf("другой пользователь уже добавил этот заказ")
	}

	// записи с заказом еще нет - добавляем
	query = `INSERT INTO orders (user_id, num, status, accrual, uploaded_at)
			  VALUES ($1, $2, $3, $4, now()) RETURNING id`

	status := constants.OrderInternalError

	err = p.storage.retryExec(ctx, query, userID, number, constants.OrderNew, 0)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.UniqueViolation == pgErr.Code {
				status = constants.OrderAlreadyUpload
			}
		}
		return status, fmt.Errorf("OrderRepo Create: %w", err)
	}

	query = `SELECT currval('orders_id_seq')`
	row = p.storage.db.QueryRowContext(ctx, query)

	var oid int64
	err = row.Scan(&oid)
	if err != nil {
		return status, fmt.Errorf("OrderRepo Create | GetAutoInc: %w", err)
	}
	if oid == 0 {
		fmt.Print("")
	}
	return constants.OrderAccepted, nil
}

func (p *OrderRepo) List(ctx context.Context, userID int64) ([]OrderRow, int, error) {
	orders := make([]OrderRow, 0)

	query := `SELECT id, user_id, num, status, accrual, uploaded_at 
			  FROM orders WHERE user_id = $1 
			  ORDER BY uploaded_at DESC`
	rows, err := p.storage.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, constants.OrderInternalError, fmt.Errorf("OrderRepo List: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var o OrderRow
		err = rows.Scan(&o.ID, &o.UserID, &o.Num, &o.Status, &o.Accrual, &o.UploadedAt)
		if err != nil {
			return nil, constants.OrderInternalError, fmt.Errorf("OrderRepo rows.Next: %w", err)
		}

		orders = append(orders, o)
	}

	if err = rows.Err(); err != nil {
		return nil, constants.OrderInternalError, fmt.Errorf("OrderRepo rows.Err: %w", err)
	}

	return orders, constants.OrdersListOk, nil
}

func (p *OrderRepo) GetUnchecked(ctx context.Context) ([]OrderRow, error) {
	orders := make([]OrderRow, 0)

	query := `SELECT id, user_id, num, status, accrual, uploaded_at 
			  FROM orders WHERE status = $1 OR status = $2`
	rows, err := p.storage.db.QueryContext(ctx, query, constants.OrderNew, constants.OrderProcessing)
	if err != nil {
		return nil, fmt.Errorf("GetUnchecked error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var o OrderRow
		err = rows.Scan(&o.ID, &o.UserID, &o.Num, &o.Status, &o.Accrual, &o.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("GetUnchecked rows.Next: %w", err)
		}

		orders = append(orders, o)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetUnchecked rows.Err: %w", err)
	}

	return orders, nil
}

func (p *OrderRepo) GetOrderByNumber(ctx context.Context, orderNumber int64) (OrderRow, error) {
	var order OrderRow

	query := `SELECT id, user_id, num, status, accrual, uploaded_at 
			  FROM orders WHERE num = $1`
	row := p.storage.db.QueryRowContext(ctx, query, orderNumber)

	err := row.Scan(&order.ID, &order.UserID, &order.Num, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		return OrderRow{}, fmt.Errorf("GetOrder Scan: %w", err)
	}

	return order, nil
}

func (p *OrderRepo) UpdateStatus(ctx context.Context, orderNumber int64, orderStatus string) error {

	query := `UPDATE orders SET status = &1 
			  WHERE num = $2`

	err := p.storage.retryExec(ctx, query, orderNumber, orderStatus)
	if err != nil {
		logger.Log().Error(fmt.Sprintf("order %v is not update", orderNumber))
		return fmt.Errorf("order is not update")
	}

	return nil
}
