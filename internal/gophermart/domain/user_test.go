package domain

import (
	"context"
	"github.com/dnsoftware/gophermart2/internal/constants"
	mock_domain "github.com/dnsoftware/gophermart2/internal/gophermart/domain/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestPasswordHash(t *testing.T) {
	hash := PassHash("pass#$%word")

	assert.Equal(t, "59e5bf62c83b5b521dc91f5baff2eb17215b43e877739571df4cd80f1b3d9d29", hash)
}

func TestLoginValidate(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "Positive",
			value: "login",
			want:  true,
		},
		{
			name:  "Negative",
			value: "lo",
			want:  false,
		},
	}

	for _, test := range tests {
		ok := loginValidate(test.value)
		assert.Equal(t, test.want, ok)
	}

}

func TestPasswordValidate(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "Пароль должен иметь заглавные, цифры, буквы",
			value: "fsYsdsds6dd",
			want:  true,
		},
		{
			name:  "Нет заглавных",
			value: "looooosttsy",
			want:  false,
		},
		{
			name:  "Слишком короткий",
			value: "s9Y",
			want:  false,
		},
	}

	for _, test := range tests {
		ok := passwordValidate(test.value)
		assert.Equal(t, test.want, ok, "%v: %v", test.name, ok)
	}

}

func TestAddUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_domain.NewMockUserStorage(ctrl)

	ctx := context.Background()

	// positive
	login := "user"
	password := "f7H456789"
	passCrypted := PassHash(password)
	m.EXPECT().Create(ctx, login, passCrypted).Return(int64(1), constants.RegisterOk, nil)

	userModel := NewUserModel(m)
	token, status, err := userModel.AddUser(ctx, login, password)

	partsToken := strings.Split(token, ".")

	require.Equal(t, 3, len(partsToken), "должен быть JWT токен")
	require.Equal(t, constants.RegisterOk, status, "Неверный статус")
	require.NoError(t, err)

	// negative короткий логин
	login = "us"
	password = "f7H456789"
	passCrypted = PassHash(password)
	m.EXPECT().Create(ctx, login, passCrypted).Return(int64(0), constants.RegisterOk, nil).AnyTimes()

	userModel = NewUserModel(m)
	_, _, err = userModel.AddUser(ctx, login, password)

	require.Error(t, err, "Должна быть ошибка длины логина")

	// negative некорректный пароль
	login = "usddd"
	password = "f7456789"
	passCrypted = PassHash(password)
	m.EXPECT().Create(ctx, login, passCrypted).Return(int64(0), constants.RegisterOk, nil).AnyTimes()

	userModel = NewUserModel(m)
	_, _, err = userModel.AddUser(ctx, login, password)

	require.Error(t, err, "Должна быть ошибка некорретный пароль")

}

func TestLoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_domain.NewMockUserStorage(ctrl)

	ctx := context.Background()

	// positive
	login := "user"
	password := "f7H456789"
	passCrypted := PassHash(password)
	m.EXPECT().FindByLoginPassword(ctx, login, passCrypted).Return(int64(1), constants.LoginOk, nil)

	userModel := NewUserModel(m)
	token, status, err := userModel.LoginUser(ctx, login, password)

	partsToken := strings.Split(token, ".")

	require.Equal(t, 3, len(partsToken), "должен быть JWT токен")
	require.Equal(t, constants.LoginOk, status, "Неверный статус")
	require.NoError(t, err)

}
