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
	_ "github.com/jackc/pgx/v5/stdlib"
	"strings"
	"time"
)

type MartStorage struct {
	db *sql.DB
}

func NewMartStorage(dsn string) (*MartStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBContextTimeout)
	defer cancel()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Log().Error(err.Error())
		return nil, err
	}

	ps := &MartStorage{
		db: db,
	}

	// создание таблиц, если не существуют
	err = ps.createDatabaseTables(ctx)
	if err != nil {
		return nil, err
	}

	return ps, nil
}

// формирование структуры БД
func (p *MartStorage) createDatabaseTables(ctx context.Context) error {
	var query string

	// users
	query = `CREATE TABLE IF NOT EXISTS users
			(
			    id SERIAL PRIMARY KEY,
			    login character varying(128) UNIQUE NOT NULL,
			    password character varying(64) NOT NULL,
			    updated_at timestamp with time zone NOT NULL
			)`

	err := p.retryExec(ctx, query)
	if err != nil {
		return err
	}

	// orders
	query = `CREATE TABLE IF NOT EXISTS orders
			(
			    id SERIAL PRIMARY KEY,
			    user_id integer NOT NULL, 
			    num bigint UNIQUE NOT NULL,
			    status character varying(16) NOT NULL,
    			accrual numeric(10,2) DEFAULT 0,
			    uploaded_at timestamp with time zone NOT NULL
			);  

			CREATE INDEX IF NOT EXISTS orders_user_id_index
				ON orders (user_id);
			CREATE INDEX IF NOT EXISTS orders_status_index
				ON orders (status);
			CREATE INDEX IF NOT EXISTS orders_uploaded_at_index
				ON orders (uploaded_at);`

	err = p.retryExec(ctx, query)
	if err != nil {
		return err
	}

	// balance
	query = `CREATE TABLE IF NOT EXISTS balances
			(
			    id SERIAL PRIMARY KEY,
			    user_id integer NOT NULL, 
			    order_number bigint NOT NULL,
    			amount numeric(10,2) DEFAULT 0,
			    processed_at timestamp with time zone NOT NULL
			); 

			CREATE INDEX IF NOT EXISTS balances_user_id_index
				ON balances (user_id);
			CREATE INDEX IF NOT EXISTS balances_order_number_index
				ON balances (order_number);
			CREATE INDEX IF NOT EXISTS balances_processed_at_index
				ON balances (processed_at);`

	err = p.retryExec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (p *MartStorage) retryExec(ctx context.Context, query string, args ...any) error {
	durations := strings.Split(constants.HTTPAttemtPeriods, ",")

	_, err := p.db.ExecContext(ctx, query, args...)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
		for _, duration := range durations {
			d, _ := time.ParseDuration(duration)
			time.Sleep(d)

			_, err = p.db.ExecContext(ctx, query, args...)
			if err == nil {
				break
			}
		}

		if err != nil {
			return fmt.Errorf("retryExec | ConnectionException: %w", err)
		}
	}

	if err != nil {
		return fmt.Errorf("retryExec: %w", err)
	}

	return nil
}
