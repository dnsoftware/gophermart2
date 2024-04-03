package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dnsoftware/gophermart2/internal/constants"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepo struct {
	storage *MartStorage
}

type UserRow struct {
	ID       int64
	Login    string
	Password string
}

func NewUserRepo(storage *MartStorage) *UserRepo {

	repo := UserRepo{
		storage: storage,
	}

	return &repo
}

// Создание пользователя
// возвращает вторым параметром статус-код операции
func (p *UserRepo) Create(ctx context.Context, login string, password string) (int64, int, error) {

	query := `INSERT INTO users (login, password, updated_at)
			  VALUES ($1, $2, now()) RETURNING id`

	status := constants.RegisterInternalError

	err := p.storage.retryExec(ctx, query, login, password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.UniqueViolation == pgErr.Code {
				status = constants.RegisterLoginExist
			}
		}
		return 0, status, fmt.Errorf("UserRepo Create: %w", err)
	}

	query = `SELECT currval('users_id_seq')`
	row := p.storage.db.QueryRowContext(ctx, query)

	var id int64

	err = row.Scan(&id)
	if err != nil {
		return 0, status, fmt.Errorf("UserRepo Create | GetAutoInc: %w", err)
	}

	return id, constants.RegisterOk, nil
}

func (p *UserRepo) FindByID(ctx context.Context, id int64) (UserRow, error) {

	item := UserRow{}

	return item, nil
}

func (p *UserRepo) FindByLoginPassword(ctx context.Context, login string, password string) (int64, int, error) {

	query := `SELECT id FROM users WHERE login = $1 AND password = $2`
	row := p.storage.db.QueryRowContext(ctx, query, login, password)

	var id int64

	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, constants.LoginBadPair, fmt.Errorf("FindByLoginPassword ErrNoRows: %w", err)
		}

		return 0, constants.LoginInternalError, fmt.Errorf("FindByLoginPassword Scan: %w", err)
	}

	return id, constants.LoginOk, nil
}
