package constants

const (
	RegisterOk = iota
	RegisterBadFormat
	RegisterLoginExist
	RegisterInternalError

	LoginOk
	LoginBadFormat
	LoginBadPair
	LoginInternalError

	OrderOk
	OrderAccepted
	OrderBadQueryFormat
	OrderNoAuth
	OrderAlreadyUpload
	OrderBadNumberFormat
	OrderInternalError

	OrdersListOk
	OrdersListNoData
	OrdersListNoAuth
	OrdersListInternalError

	BalanceOk
	BalanceNoAuth
	BalanceInternalError

	WithdrawOk
	WithdrawNoAuth
	WithdrawNotEnoughFunds
	WithdrawBadOrderNumber
	WithdrawInternalError

	WithdrawalsOk
	WithdrawalsNoOne
	WithdrawalsNoAuth
	WithdrawalsInternalError

	AccrualOk
	AccrualNoReg
	AccrualNumberExceeded
	AccrualInternalError

	Unknown
)

func StatusData(code int) (int, string) {
	switch code {

	case RegisterOk:
		return 200, "пользователь успешно зарегистрирован и аутентифицирован"
	case RegisterBadFormat:
		return 400, StatusBadRequestFormat
	case RegisterLoginExist:
		return 409, "логин уже занят"
	case RegisterInternalError:
		return 500, StatusInternalServerError

	case LoginOk:
		return 200, "пользователь успешно аутентифицирован"
	case LoginBadFormat:
		return 400, StatusBadRequestFormat
	case LoginBadPair:
		return 401, "неверная пара логин/пароль"
	case LoginInternalError:
		return 500, StatusInternalServerError

	case OrderOk:
		return 200, "номер заказа уже был загружен этим пользователем"
	case OrderAccepted:
		return 202, "новый номер заказа принят в обработку"
	case OrderBadQueryFormat:
		return 400, StatusBadRequestFormat
	case OrderNoAuth:
		return 401, "пользователь не аутентифицирован"
	case OrderAlreadyUpload:
		return 409, "номер заказа уже был загружен другим пользователем"
	case OrderBadNumberFormat:
		return 422, StatusBadNumberFormat
	case OrderInternalError:
		return 500, StatusInternalServerError

	case OrdersListOk:
		return 200, StatusSuccessfulRequest
	case OrdersListNoData:
		return 204, "нет данных для ответа"
	case OrdersListNoAuth:
		return 401, StatusUnauthorized
	case OrdersListInternalError:
		return 500, StatusInternalServerError

	case BalanceOk:
		return 200, StatusSuccessfulRequest
	case BalanceNoAuth:
		return 401, StatusUnauthorized
	case BalanceInternalError:
		return 500, StatusInternalServerError

	case WithdrawOk:
		return 200, StatusSuccessfulRequest
	case WithdrawNoAuth:
		return 401, StatusUnauthorized
	case WithdrawNotEnoughFunds:
		return 402, "на счету недостаточно средств"
	case WithdrawBadOrderNumber:
		return 422, StatusBadNumberFormat
	case WithdrawInternalError:
		return 500, StatusInternalServerError

	case WithdrawalsOk:
		return 200, StatusSuccessfulRequest
	case WithdrawalsNoOne:
		return 204, "нет ни одного списания"
	case WithdrawalsNoAuth:
		return 401, StatusUnauthorized
	case WithdrawalsInternalError:
		return 500, StatusInternalServerError

	case AccrualOk:
		return 200, StatusSuccessfulRequest
	case AccrualNoReg:
		return 204, "заказ не зарегистрирован в системе расчёта"
	case AccrualNumberExceeded:
		return 429, "превышено количество запросов к сервису"
	case AccrualInternalError:
		return 500, StatusInternalServerError

	case Unknown:
		return 1000, "unknown 1000"

	default:
		return 2000, "unknown 2000"
	}
}
