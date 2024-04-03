package domain

import (
	"context"
	"errors"
	"fmt"
	"github.com/dnsoftware/gophermart2/internal/constants"
	mock_domain "github.com/dnsoftware/gophermart2/internal/gophermart/domain/mocks"
	"github.com/dnsoftware/gophermart2/internal/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBalance_UserBalance(t *testing.T) {
	type fields struct {
		storage BalanceStorage
	}
	type args struct {
		ctx    context.Context
		userID int64
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalance := mock_domain.NewMockBalanceStorage(ctrl)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CurrentBalance
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "Positive",
			fields: fields{storage: mockBalance},
			args: args{
				ctx:    ctx,
				userID: int64(1),
			},
			want: func() *CurrentBalance {
				mockBalance.EXPECT().GetUserBalance(ctx, int64(1)).Return(float32(100), nil)
				mockBalance.EXPECT().GetUserWithdrawn(ctx, int64(1)).Return(float32(50), nil)

				return &CurrentBalance{
					Current:   100,
					Withdrawn: 50,
				}
			}(),
			wantErr: assert.NoError,
		},
		{
			name:   "Bad balance",
			fields: fields{storage: mockBalance},
			args: args{
				ctx:    ctx,
				userID: int64(1),
			},
			want: func() *CurrentBalance {
				mockBalance.EXPECT().GetUserBalance(ctx, int64(1)).Return(float32(0), errors.New("ошибка получения баланса")).AnyTimes()
				mockBalance.EXPECT().GetUserWithdrawn(ctx, int64(1)).Return(float32(50), nil).AnyTimes()

				return nil
			}(),
			wantErr: assert.Error,
		},
		{
			name:   "Bad withdraw",
			fields: fields{storage: mockBalance},
			args: args{
				ctx:    ctx,
				userID: int64(1),
			},
			want: func() *CurrentBalance {
				mockBalance.EXPECT().GetUserBalance(ctx, int64(1)).Return(float32(100), nil).AnyTimes()
				mockBalance.EXPECT().GetUserWithdrawn(ctx, int64(1)).Return(float32(0), errors.New("ошибка получения списаний")).AnyTimes()

				return nil
			}(),
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Balance{
				storage: tt.fields.storage,
			}
			got, err := b.UserBalance(tt.args.ctx, tt.args.userID)
			if !tt.wantErr(t, err, fmt.Sprintf("UserBalance(%v, %v)", tt.args.ctx, tt.args.userID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "UserBalance(%v, %v)", tt.args.ctx, tt.args.userID)
		})
	}
}

func TestBalance_UserWithrawalsList(t *testing.T) {
	type fields struct {
		storage BalanceStorage
	}
	type args struct {
		ctx    context.Context
		userID int64
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalance := mock_domain.NewMockBalanceStorage(ctrl)
	tm, _ := time.Parse(time.RFC3339, "2024-03-19T15:24:39-07:00")
	mockBalance.EXPECT().GetUserWithdrawList(ctx, int64(1)).Return([]storage.WithdrawRow{{
		Order:       3840576627,
		Sum:         111,
		ProcessedAt: tm,
	}}, nil)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []WithdrawItem
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "Positive",
			fields: fields{storage: mockBalance},
			args: args{
				ctx:    ctx,
				userID: int64(1),
			},
			want: []WithdrawItem{{
				Order:       "3840576627",
				Sum:         111,
				ProcessedAt: "2024-03-19T15:24:39-07:00",
			}},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Balance{
				storage: tt.fields.storage,
			}
			got, err := b.UserWithrawalsList(tt.args.ctx, tt.args.userID)
			if !tt.wantErr(t, err, fmt.Sprintf("UserWithrawalsList(%v, %v)", tt.args.ctx, tt.args.userID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "UserWithrawalsList(%v, %v)", tt.args.ctx, tt.args.userID)
		})
	}
}

func TestBalance_Withraw(t *testing.T) {
	type fields struct {
		storage BalanceStorage
	}
	type args struct {
		ctx    context.Context
		userID int64
		number int64
		amount float32
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalance := mock_domain.NewMockBalanceStorage(ctrl)

	tests := []struct {
		name    string
		fields  fields
		prepare func()
		args    args
		want    int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "Positive",
			fields: fields{storage: mockBalance},
			prepare: func() {
				mockBalance.EXPECT().GetUserBalance(ctx, int64(1)).Return(float32(1000), nil)
				mockBalance.EXPECT().GetUserWithdrawn(ctx, int64(1)).Return(float32(50), nil)
				mockBalance.EXPECT().WithdrawTransaction(ctx, int64(1), int64(3840576627), float32(729.98)).Return(nil)
			},
			args: args{
				ctx:    ctx,
				userID: int64(1),
				number: 3840576627,
				amount: 729.98,
			},
			want:    constants.WithdrawalsOk,
			wantErr: assert.NoError,
		},
		{
			name:   "Low balance",
			fields: fields{storage: mockBalance},
			prepare: func() {
				mockBalance.EXPECT().GetUserBalance(ctx, int64(1)).Return(float32(700), nil)
				mockBalance.EXPECT().GetUserWithdrawn(ctx, int64(1)).Return(float32(50), nil)
				mockBalance.EXPECT().WithdrawTransaction(ctx, int64(1), int64(3840576627), float32(729.98)).Return(nil).AnyTimes()
			},

			args: args{
				ctx:    ctx,
				userID: int64(1),
				number: 3840576627,
				amount: 729.98,
			},
			want:    constants.WithdrawNotEnoughFunds,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Balance{
				storage: tt.fields.storage,
			}
			tt.prepare()
			got, err := b.Withraw(tt.args.ctx, tt.args.userID, tt.args.number, tt.args.amount)
			if !tt.wantErr(t, err, fmt.Sprintf("Withraw(%v, %v, %v, %v)", tt.args.ctx, tt.args.userID, tt.args.number, tt.args.amount)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Withraw(%v, %v, %v, %v)", tt.args.ctx, tt.args.userID, tt.args.number, tt.args.amount)
		})
	}
}
