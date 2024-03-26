package domain

//
//import (
//	"github.com/dnsoftware/gophermart2/internal/gophermart/config"
//	"github.com/dnsoftware/gophermart2/internal/logger"
//)
//
//type Gophermart struct {
//	cfg     *config.Config
//	User    *User
//	Order   *Order
//	Balance *Balance
//	Accrual *Accrual
//}
//
//func NewGophermart(config *config.Config,
//	userStorage UserStorage,
//	orderStorage OrderStorage,
//	balanceStorage BalanceStorage,
//	accrualStorage AccrualStorage,
//) (*Gophermart, error) {
//
//	user, err := NewUserModel(userStorage)
//	if err != nil {
//		logger.Log().Error("Error creating UserModel: " + err.Error())
//		return nil, err
//	}
//
//	order, err := NewOrderModel(orderStorage)
//	if err != nil {
//		logger.Log().Error("Error creating OrderModel: " + err.Error())
//		return nil, err
//	}
//
//	balance, err := NewBalanceModel(balanceStorage)
//	if err != nil {
//		logger.Log().Error("Error creating OrderModel: " + err.Error())
//		return nil, err
//	}
//
//	accrual, err := NewAccrualModel(accrualStorage)
//	if err != nil {
//		logger.Log().Error("Error creating AccrualModel: " + err.Error())
//		return nil, err
//	}
//
//	gophermart := &Gophermart{
//		cfg:     config,
//		User:    user,
//		Order:   order,
//		Balance: balance,
//		Accrual: accrual,
//	}
//
//	return gophermart, nil
//}
