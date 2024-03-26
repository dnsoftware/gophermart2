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
		return 400, "неверный формат запроса"
	case RegisterLoginExist:
		return 409, "логин уже занят"
	case RegisterInternalError:
		return 500, "внутренняя ошибка сервера"

	case LoginOk:
		return 200, "пользователь успешно аутентифицирован"
	case LoginBadFormat:
		return 400, "неверный формат запроса"
	case LoginBadPair:
		return 401, "неверная пара логин/пароль"
	case LoginInternalError:
		return 500, "внутренняя ошибка сервера"

	case OrderOk:
		return 200, "номер заказа уже был загружен этим пользователем"
	case OrderAccepted:
		return 202, "новый номер заказа принят в обработку"
	case OrderBadQueryFormat:
		return 400, "неверный формат запроса"
	case OrderNoAuth:
		return 401, "пользователь не аутентифицирован"
	case OrderAlreadyUpload:
		return 409, "номер заказа уже был загружен другим пользователем"
	case OrderBadNumberFormat:
		return 422, "неверный формат номера заказа"
	case OrderInternalError:
		return 500, "внутренняя ошибка сервера"

	case OrdersListOk:
		return 200, "успешная обработка запроса"
	case OrdersListNoData:
		return 204, "нет данных для ответа"
	case OrdersListNoAuth:
		return 401, "пользователь не авторизован"
	case OrdersListInternalError:
		return 500, "внутренняя ошибка сервера"

	case BalanceOk:
		return 200, "успешная обработка запроса"
	case BalanceNoAuth:
		return 401, "пользователь не авторизован"
	case BalanceInternalError:
		return 500, "внутренняя ошибка сервера"

	case WithdrawOk:
		return 200, "успешная обработка запроса"
	case WithdrawNoAuth:
		return 401, "пользователь не авторизован"
	case WithdrawNotEnoughFunds:
		return 402, "на счету недостаточно средств"
	case WithdrawBadOrderNumber:
		return 422, "неверный номер заказа"
	case WithdrawInternalError:
		return 500, "внутренняя ошибка сервера"

	case WithdrawalsOk:
		return 200, "успешная обработка запроса"
	case WithdrawalsNoOne:
		return 204, "нет ни одного списания"
	case WithdrawalsNoAuth:
		return 401, "пользователь не авторизован"
	case WithdrawalsInternalError:
		return 500, "внутренняя ошибка сервера"

	case AccrualOk:
		return 200, "успешная обработка запроса"
	case AccrualNoReg:
		return 204, "заказ не зарегистрирован в системе расчёта"
	case AccrualNumberExceeded:
		return 429, "превышено количество запросов к сервису"
	case AccrualInternalError:
		return 500, "внутренняя ошибка сервера"

	case Unknown:
		return 1000, "unknown 1000"

	default:
		return 2000, "unknown 2000"
	}
}
