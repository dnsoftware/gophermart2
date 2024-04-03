package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dnsoftware/gophermart2/internal/constants"
	"github.com/dnsoftware/gophermart2/internal/gophermart/domain"
	"github.com/dnsoftware/gophermart2/internal/logger"
	"github.com/go-chi/chi/v5"
	"net/http"
	"regexp"
	"strconv"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()
	return r
}

// регистрация пользователя
func (h *Server) userRegister(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	var buf bytes.Buffer

	var user domain.UserItem

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &user); err != nil {
		code, message := constants.StatusData(constants.RegisterBadFormat)
		http.Error(res, message, code)
		return
	}

	token, status, err := h.userMart.AddUser(ctx, user.Login, user.Password)
	if err != nil {
		logger.Log().Error(err.Error())
		code, message := constants.StatusData(status)
		http.Error(res, message+", "+err.Error(), code)
		return
	}

	if status != constants.RegisterOk {
		code, message := constants.StatusData(status)
		http.Error(res, message, code)
		return
	}

	res.Header().Set("Content-Type", constants.ApplicationJSON)
	bearer := "Bearer " + token
	res.Header().Set(constants.HeaderAuthorization, bearer)
	code, message := constants.StatusData(status)
	res.WriteHeader(code)
	res.Write([]byte(message))

}

// вход пользователя
func (h *Server) userLogin(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	var buf bytes.Buffer

	var user domain.UserItem

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &user); err != nil {
		code, message := constants.StatusData(constants.LoginBadFormat)
		http.Error(res, message, code)
		return
	}

	token, status, err := h.userMart.LoginUser(ctx, user.Login, user.Password)
	if err != nil {
		logger.Log().Error(err.Error())
		code, message := constants.StatusData(status)
		http.Error(res, message+", "+err.Error(), code)
		return
	}

	if status != constants.LoginOk {
		code, message := constants.StatusData(status)
		http.Error(res, message, code)
		return
	}

	res.Header().Set("Content-Type", constants.ApplicationJSON)
	bearer := "Bearer " + token
	res.Header().Set(constants.HeaderAuthorization, bearer)
	code, message := constants.StatusData(status)
	res.WriteHeader(code)
	res.Write([]byte(message))

}

func (h *Server) userOrderUpload(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	var buf bytes.Buffer

	uid := ctx.Value(constants.UserIDKey)
	userID, ok := uid.(int64)
	if !ok {
		http.Error(res, "", http.StatusUnauthorized)
		return
	}

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	number := buf.String()
	re := regexp.MustCompile(`^\d+$`)
	if !re.MatchString(number) {
		http.Error(res, "", http.StatusBadRequest)
		return
	}

	orderID, _ := strconv.ParseInt(number, 10, 64)

	status, err := h.orderMart.AddOrder(ctx, userID, orderID)
	if err != nil {
		code, message := constants.StatusData(status)
		http.Error(res, message, code)
		return
	}

	res.Header().Set("Content-Type", constants.ApplicationJSON)
	code, message := constants.StatusData(status)
	res.WriteHeader(code)
	res.Write([]byte(message))
}

func (h *Server) userOrdersList(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	uid := ctx.Value(constants.UserIDKey)
	userID, ok := uid.(int64)
	if !ok {
		http.Error(res, "", http.StatusUnauthorized)
		return
	}

	list, status, err := h.orderMart.OrdersList(ctx, userID)
	if err != nil {
		code, message := constants.StatusData(status)
		http.Error(res, message, code)
		return
	}

	body, err := json.Marshal(list)
	if err != nil {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", constants.ApplicationJSON)
	code, _ := constants.StatusData(status)
	res.WriteHeader(code)
	res.Write(body)
}

func (h *Server) userBalance(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	uid := ctx.Value(constants.UserIDKey)
	userID, ok := uid.(int64)
	if !ok {
		http.Error(res, "", http.StatusUnauthorized)
		return
	}

	currentBalance, err := h.balanceMart.UserBalance(ctx, userID)
	if err != nil {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(currentBalance)
	if err != nil {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", constants.ApplicationJSON)
	res.WriteHeader(http.StatusOK)
	res.Write(body)
}

func (h *Server) userWithdrawals(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	uid := ctx.Value(constants.UserIDKey)
	userID, ok := uid.(int64)
	if !ok {
		http.Error(res, "", http.StatusUnauthorized)
		return
	}

	withdrawalsList, err := h.balanceMart.UserWithrawalsList(ctx, userID)
	if err != nil {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(withdrawalsList)
	if err != nil {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", constants.ApplicationJSON)

	if withdrawalsList != nil {
		res.WriteHeader(http.StatusOK)
	} else {
		res.WriteHeader(http.StatusNoContent)
	}

	res.Write(body)
}

func (h *Server) userWithdraw(res http.ResponseWriter, req *http.Request) {
	type wReq struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}
	ctx, cancel := context.WithTimeout(req.Context(), constants.DBContextTimeout)
	defer cancel()

	uid := ctx.Value(constants.UserIDKey)
	userID, ok := uid.(int64)
	if !ok {
		http.Error(res, "", http.StatusUnauthorized)
		return
	}

	var buf bytes.Buffer
	var reqData wReq

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &reqData); err != nil {
		http.Error(res, "", http.StatusInternalServerError)
		return
	}

	orderNum, _ := strconv.ParseInt(reqData.Order, 10, 64)
	status, err := h.balanceMart.Withraw(ctx, userID, orderNum, reqData.Sum)
	if err != nil {
		logger.Log().Error(err.Error())
	}

	var httpStatus int
	switch status {
	case constants.WithdrawBadOrderNumber:
		httpStatus = http.StatusUnprocessableEntity
	case constants.WithdrawInternalError:
		httpStatus = http.StatusInternalServerError
	case constants.WithdrawNotEnoughFunds:
		httpStatus = http.StatusPaymentRequired
	case constants.WithdrawalsOk:
		httpStatus = http.StatusOK
	}

	res.Header().Set("Content-Type", constants.ApplicationJSON)
	res.WriteHeader(httpStatus)

}
