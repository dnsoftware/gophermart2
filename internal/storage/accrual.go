package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dnsoftware/gophermart2/internal/constants"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type AccrualRepo struct {
	client                *http.Client
	orderEndpointTemplate string
}

type AccrualRow struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

func NewAccrualRepo(address string) *AccrualRepo {
	return &AccrualRepo{
		client:                &http.Client{},
		orderEndpointTemplate: address + constants.AccrualOrderEndpoint,
	}
}

// Получить данные по заказу
func (a *AccrualRepo) GetOrder(orderNum int64) (*AccrualRow, int, error) {
	ctx := context.Background()
	buf := &bytes.Buffer{}

	url := a.orderEndpoint(orderNum)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, buf)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	request.Header.Set("Content-Type", constants.ApplicationJSON)

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	row := &AccrualRow{}

	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(data, row)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		errMessage := string(data) + ", Retry-After: " + resp.Header.Get("Retry-After")
		return nil, http.StatusTooManyRequests, fmt.Errorf(errMessage)
	}

	return row, resp.StatusCode, nil
}

func (a *AccrualRepo) orderEndpoint(orderNum int64) string {
	strNum := strconv.FormatInt(orderNum, 10)

	return strings.Replace(a.orderEndpointTemplate, "{number}", strNum, -1)
}
