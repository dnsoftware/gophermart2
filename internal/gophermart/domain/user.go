package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dnsoftware/gophermart2/internal/constants"
	"github.com/dnsoftware/gophermart2/internal/logger"
	"github.com/dnsoftware/gophermart2/internal/storage"
	"unicode"
	"unicode/utf8"
)

type UserStorage interface {
	Create(ctx context.Context, login string, password string) (int64, int, error)
	FindByID(ctx context.Context, id int64) (storage.UserRow, error)
	FindByLoginPassword(ctx context.Context, login string, password string) (int64, int, error)
}

type User struct {
	storage UserStorage
}

// UserItem plain structure
type UserItem struct {
	ID       int64
	Login    string `json:"login"`
	Password string `json:"password"`
}

func NewUserModel(storage UserStorage) *User {
	user := &User{
		storage: storage,
	}

	return user
}

// возвращает токен нового пользователя, статус завершения операции и ошибку
func (u *User) AddUser(ctx context.Context, login string, password string) (string, int, error) {
	passCrypted := PassHash(password)

	status := constants.RegisterInternalError

	if !loginValidate(login) {
		status = constants.RegisterBadFormat
		return "", status, fmt.Errorf("некорректный логин, длина должна быть больше/равна %v", constants.MinLoginLength)
	}
	if !passwordValidate(password) {
		status = constants.RegisterBadFormat
		return "", status, fmt.Errorf("некорректный пароль, должен иметь заглавные, цифры, буквы и быть определенной длины >= %v", constants.MinPasswordLength)
	}

	id, status, err := u.storage.Create(ctx, login, passCrypted)
	if err != nil {
		logger.Log().Error("AddUser: " + err.Error())
		return "", status, err
	}

	token, err := BuildJWTString(id)
	if err != nil {
		return "", constants.RegisterInternalError, fmt.Errorf("JWT error")
	}

	return token, status, nil
}

// возвращает токен пользователя, статус завершения операции и ошибку
func (u *User) LoginUser(ctx context.Context, login string, password string) (string, int, error) {
	passCrypted := PassHash(password)

	status := constants.LoginInternalError

	if !loginValidate(login) {
		status = constants.LoginBadFormat
		return "", status, fmt.Errorf("некорректный логин, длина должна быть больше/равна %v", constants.MinLoginLength)
	}

	id, status, err := u.storage.FindByLoginPassword(ctx, login, passCrypted)
	if err != nil {
		logger.Log().Error("LoginUser: " + err.Error())
		return "", status, err
	}

	token, err := BuildJWTString(id)
	if err != nil {
		return "", constants.LoginInternalError, fmt.Errorf("JWT error")
	}

	return token, status, nil
}

func PassHash(password string) string {
	data := sha256.Sum256([]byte(password + constants.PasswordSalt))

	return hex.EncodeToString(data[:])
}

// Логин должен быть определенной длины
func loginValidate(login string) bool {

	length := utf8.RuneCountInString(login)

	return length >= constants.MinLoginLength
}

// Пароль должен иметь заглавные, цифры, спецсимволы, буквы и быть определенной длины
/*
возможное доп.требование - наличие спецсимвола
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
*/
func passwordValidate(password string) bool {

	number, upper, letter := false, false, false
	length := 0

	for _, c := range password {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsLetter(c) || c == ' ':
			letter = true
		default:
			return false
		}
		length++
	}

	return (length >= constants.MinPasswordLength) && number && upper && letter
}
