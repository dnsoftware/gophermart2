package domain

import (
	"context"
	"fmt"
	"github.com/dnsoftware/gophermart2/internal/constants"
	mock_domain "github.com/dnsoftware/gophermart2/internal/gophermart/domain/mocks"
	"github.com/dnsoftware/gophermart2/internal/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestOrder_AddOrder(t *testing.T) {
	type fields struct {
		storage       OrderStorage
		ordersToCheck UncheckedOrders
		ordersToSave  CheckedOrders
		balanceAdd    BalanceAdd
	}
	type args struct {
		ctx    context.Context
		userID int64
		number int64
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	chanUnchecked := NewOrdersUnchecked()
	chanChecked := NewOrdersChecked()

	mockOrder := mock_domain.NewMockOrderStorage(ctrl)
	mockOrder.EXPECT().Create(ctx, int64(1), int64(3840576627)).Return(constants.OrderAccepted, nil)

	mockBalance := mock_domain.NewMockBalanceAdd(ctrl)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Positive",
			fields: fields{
				storage:       mockOrder,
				ordersToCheck: chanUnchecked,
				ordersToSave:  chanChecked,
				balanceAdd:    mockBalance,
			},
			args:    args{ctx, 1, 3840576627},
			want:    constants.OrderAccepted,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{
				storage:       tt.fields.storage,
				ordersToCheck: tt.fields.ordersToCheck,
				ordersToSave:  tt.fields.ordersToSave,
				balanceAdd:    tt.fields.balanceAdd,
			}
			got, err := o.AddOrder(tt.args.ctx, tt.args.userID, tt.args.number)
			if !tt.wantErr(t, err, fmt.Sprintf("AddOrder(%v, %v, %v)", tt.args.ctx, tt.args.userID, tt.args.number)) {
				return
			}
			assert.Equalf(t, tt.want, got, "AddOrder(%v, %v, %v)", tt.args.ctx, tt.args.userID, tt.args.number)
		})
	}
}

func TestOrder_SetStatus(t *testing.T) {
	type fields struct {
		storage       OrderStorage
		ordersToCheck UncheckedOrders
		ordersToSave  CheckedOrders
		balanceAdd    BalanceAdd
	}
	type args struct {
		ctx         context.Context
		orderNumber int64
		orderStatus string
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	chanUnchecked := NewOrdersUnchecked()
	chanChecked := NewOrdersChecked()

	mockOrder := mock_domain.NewMockOrderStorage(ctrl)
	mockOrder.EXPECT().UpdateStatus(ctx, int64(3840576627), constants.OrderProcessing).Return(nil)

	mockBalance := mock_domain.NewMockBalanceAdd(ctrl)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Positive",
			fields: fields{
				storage:       mockOrder,
				ordersToCheck: chanUnchecked,
				ordersToSave:  chanChecked,
				balanceAdd:    mockBalance,
			},
			args: args{
				ctx:         ctx,
				orderNumber: 3840576627,
				orderStatus: constants.OrderProcessing,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{
				storage:       tt.fields.storage,
				ordersToCheck: tt.fields.ordersToCheck,
				ordersToSave:  tt.fields.ordersToSave,
				balanceAdd:    tt.fields.balanceAdd,
			}
			tt.wantErr(t, o.SetStatus(tt.args.ctx, tt.args.orderNumber, tt.args.orderStatus), fmt.Sprintf("SetStatus(%v, %v, %v)", tt.args.ctx, tt.args.orderNumber, tt.args.orderStatus))
		})
	}
}

func TestOrder_OrdersList(t *testing.T) {
	type fields struct {
		storage       OrderStorage
		ordersToCheck UncheckedOrders
		ordersToSave  CheckedOrders
		balanceAdd    BalanceAdd
	}
	type args struct {
		ctx    context.Context
		userID int64
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	chanUnchecked := NewOrdersUnchecked()
	chanChecked := NewOrdersChecked()

	tm, _ := time.Parse(time.RFC3339, "2024-03-19T15:24:39-07:00")

	mockOrder := mock_domain.NewMockOrderStorage(ctrl)
	mockOrder.EXPECT().List(ctx, int64(1)).Return([]storage.OrderRow{{
		ID:         1,
		UserID:     1,
		Num:        3840576627,
		Status:     constants.OrderProcessed,
		Accrual:    729.98,
		UploadedAt: tm,
	}}, constants.OrdersListOk, nil)

	mockBalance := mock_domain.NewMockBalanceAdd(ctrl)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []OrderItem
		want1   int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Positive",
			fields: fields{
				storage:       mockOrder,
				ordersToCheck: chanUnchecked,
				ordersToSave:  chanChecked,
				balanceAdd:    mockBalance,
			},
			args: args{
				ctx:    ctx,
				userID: int64(1),
			},
			want: []OrderItem{{
				Number:     "3840576627",
				Status:     constants.OrderProcessed,
				Accrual:    729.98,
				UploadedAt: "2024-03-19T15:24:39-07:00",
			}},
			want1:   constants.OrdersListOk,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{
				storage:       tt.fields.storage,
				ordersToCheck: tt.fields.ordersToCheck,
				ordersToSave:  tt.fields.ordersToSave,
				balanceAdd:    tt.fields.balanceAdd,
			}
			got, got1, err := o.OrdersList(tt.args.ctx, tt.args.userID)
			if !tt.wantErr(t, err, fmt.Sprintf("OrdersList(%v, %v)", tt.args.ctx, tt.args.userID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "OrdersList(%v, %v)", tt.args.ctx, tt.args.userID)
			assert.Equalf(t, tt.want1, got1, "OrdersList(%v, %v)", tt.args.ctx, tt.args.userID)
		})
	}
}

func TestOrder_ProcessUnchecked(t *testing.T) {
	type fields struct {
		storage       OrderStorage
		ordersToCheck UncheckedOrders
		ordersToSave  CheckedOrders
		balanceAdd    BalanceAdd
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	chanUnchecked := NewOrdersUnchecked()
	chanChecked := NewOrdersChecked()

	mockOrder := mock_domain.NewMockOrderStorage(ctrl)
	tm, _ := time.Parse(time.RFC3339, "2024-03-19T15:24:39-07:00")
	mockOrder.EXPECT().GetUnchecked(ctx).Return([]storage.OrderRow{{
		ID:         1,
		UserID:     1,
		Num:        3840576627,
		Status:     constants.OrderProcessed,
		Accrual:    729.98,
		UploadedAt: tm,
	}}, nil).AnyTimes()

	mockBalance := mock_domain.NewMockBalanceAdd(ctrl)

	mockAccrual := mock_domain.NewMockAccrualStorage(ctrl)
	mockAccrual.EXPECT().GetOrder(int64(3840576627)).Return(&storage.AccrualRow{
		Order:   "3840576627",
		Status:  constants.OrderProcessed,
		Accrual: 729.98,
	}, http.StatusOK, nil).AnyTimes()
	accrual := NewAccrualModel(mockAccrual, chanUnchecked, chanChecked)
	go accrual.StartAccrualChecker(ctx)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Positive",
			fields: fields{
				storage:       mockOrder,
				ordersToCheck: chanUnchecked,
				ordersToSave:  chanChecked,
				balanceAdd:    mockBalance,
			},
			args: args{ctx},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{
				storage:       tt.fields.storage,
				ordersToCheck: tt.fields.ordersToCheck,
				ordersToSave:  tt.fields.ordersToSave,
				balanceAdd:    tt.fields.balanceAdd,
			}
			o.ProcessUnchecked(tt.args.ctx)
			number := accrual.ordersToCheck.Pop(ctx)

			require.Equal(t, int64(3840576627), number, "Номера ордеров на проверку должны совпадать")

		})
	}
}

func TestOrder_ProcessChecked(t *testing.T) {
	type fields struct {
		storage       OrderStorage
		ordersToCheck UncheckedOrders
		ordersToSave  CheckedOrders
		balanceAdd    BalanceAdd
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	chanUnchecked := NewOrdersUnchecked()
	chanChecked := NewOrdersChecked()

	mockOrder := mock_domain.NewMockOrderStorage(ctrl)
	mockBalance := mock_domain.NewMockBalanceAdd(ctrl)
	mockBalance.EXPECT().AddTransaction(ctx, int64(3840576627), float32(729.98)).Return(nil)

	mockAccrual := mock_domain.NewMockAccrualStorage(ctrl)
	mockAccrual.EXPECT().GetOrder(int64(3840576627)).Return(&storage.AccrualRow{
		Order:   "3840576627",
		Status:  "PROCESSED",
		Accrual: 729.98,
	}, http.StatusOK, nil).AnyTimes()
	accrual := NewAccrualModel(mockAccrual, chanUnchecked, chanChecked)
	go accrual.StartAccrualChecker(ctx)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Positive",
			fields: fields{
				storage:       mockOrder,
				ordersToCheck: chanUnchecked,
				ordersToSave:  chanChecked,
				balanceAdd:    mockBalance,
			},
			args: args{ctx},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{
				storage:       tt.fields.storage,
				ordersToCheck: tt.fields.ordersToCheck,
				ordersToSave:  tt.fields.ordersToSave,
				balanceAdd:    tt.fields.balanceAdd,
			}
			o.ProcessChecked(tt.args.ctx)
			accRow, status, err := accrual.storage.GetOrder(int64(3840576627))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)

			accRowInt64, _ := strconv.ParseInt(accRow.Order, 10, 64)
			accrual.ordersToSave.Push(accRowInt64, accRow.Status, accRow.Accrual)
			time.Sleep(3 * time.Second)
		})
	}
}
