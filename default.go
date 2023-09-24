package pawapay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pariz/gountries"
	"github.com/pkg/errors"
)

const (
	recipientType = "MSISDN"
)

type OperationType struct {
	OperationType string `json:"operationType"`
	Status        string `json:"status"`
}

type Correspondent struct {
	Correspondent  string          `json:"correspondent"`
	OperationTypes []OperationType `json:"operationTypes"`
}

type MomoMapping struct {
	Country        string          `json:"country"`
	Extension      string          `json:"extension"`
	Correspondents []Correspondent `json:"correspondents"`
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type PayoutRequest struct {
	PayoutId      string
	Amount        Amount
	Description   string
	PhoneNumber   PhoneNumber
	Correspondent string
}

type DepositRequest struct {
	DepositId     string
	Amount        Amount
	Description   string
	PhoneNumber   PhoneNumber
	Correspondent string
	PreAuthCode   string
}

// PhoneNumber holds country code and number, eg countryCode:234, number: 7017238745
type PhoneNumber struct {
	CountryCode string `json:"countryCode"`
	Number      string `json:"number"`
}

type Address struct {
	Value string `json:"value"`
}

type Recipient struct {
	Type    string  `json:"type"`
	Address Address `json:"address"`
}

// APIAnnotation is a representation of provider api request and response
type APIAnnotation struct {
	URL             string `json:"url"`
	RequestPayload  string `json:"requestPayload"`
	ResponsePayload string `json:"responsePayload"`
	ResponseCode    int    `json:"responseCode"`
}

type CreatePayoutRequest struct {
	PayoutId             string    `json:"payoutId"`
	Amount               string    `json:"amount"`
	Currency             string    `json:"currency"`
	Country              string    `json:"country"`
	Correspondent        string    `json:"correspondent"`
	Recipient            Recipient `json:"recipient"`
	CustomerTimestamp    string    `json:"customerTimestamp"`
	StatementDescription string    `json:"statementDescription"`
}

type Payer struct {
	Type    string  `json:"type"`
	Address Address `json:"address"`
}

type CreatePayoutResponse struct {
	PayoutID   string `json:"payoutId"`
	Status     string `json:"status"`
	Created    string `json:"created"`
	Annotation APIAnnotation
}

type CreateBulkPayoutResponse struct {
	Result     []CreatePayoutResponse
	Annotation APIAnnotation
}

type FailureReason struct {
	FailureCode    string `json:"failureCode"`
	FailureMessage string `json:"failureMessage"`
}

type Payout struct {
	Amount               string                 `json:"amount"`
	Correspondent        string                 `json:"correspondent"`
	CorrespondentIds     map[string]interface{} `json:"correspondentIds"`
	Country              string                 `json:"country"`
	Created              string                 `json:"created"`
	Currency             string                 `json:"currency"`
	CustomerTimestamp    string                 `json:"customerTimestamp"`
	PayoutID             string                 `json:"payoutId"`
	Recipient            Recipient              `json:"recipient"`
	StatementDescription string                 `json:"statementDescription"`
	Status               string                 `json:"status"`
	FailureReason        FailureReason          `json:"failureReason"`
	Annotation           APIAnnotation
}

// IsSuccessful reports if a payout is successful
func (t Payout) IsSuccessful() bool { return strings.EqualFold(t.Status, "completed") }

// IsFailed reports if a payout failed
func (t Payout) IsFailed() bool { return strings.EqualFold(t.Status, "failed") }

// IsPending reports if a payout is still pending
func (t Payout) IsPending() bool { return !t.IsSuccessful() && !t.IsFailed() && t.Status != "" }

// IsNotFound checks a payout response to see if the transaction is not found
func (t Payout) IsNotFound() bool {
	return t.Annotation.ResponseCode == 200 && t.Annotation.ResponsePayload == "[]"
}

type ResendCallbackRequest struct {
	PayoutId  string `json:"payoutId,omitempty"`
	DepositId string `json:"depositId,omitempty"`
}

type PayoutStatusResponse struct {
	PayoutId   string `json:"payoutId"`
	Status     string `json:"status"`
	Annotation APIAnnotation
}

type CreateDepositRequest struct {
	DepositId            string `json:"depositId"`
	Amount               string `json:"amount"`
	Currency             string `json:"currency"`
	Country              string `json:"country"`
	Correspondent        string `json:"correspondent"`
	Payer                Payer  `json:"payer"`
	CustomerTimestamp    string `json:"customerTimestamp"`
	StatementDescription string `json:"statementDescription"`
	PreAuthorizationCode string `json:"preAuthorisationCode"`
}

type CreateDepositResponse struct {
	DepositId  string `json:"depositId"`
	Status     string `json:"status"`
	Created    string `json:"created"`
	Annotation APIAnnotation
}

type CreateBulkDepositResponse struct {
	Result     []CreateDepositResponse
	Annotation APIAnnotation
}

// TimeProviderFunc represents a provider of time
type TimeProviderFunc func() time.Time

func (s *Service) makeRequest(method, resource string, reqBody interface{}, resp interface{}) (APIAnnotation, error) {

	URL := fmt.Sprintf("%s/%s", s.config.BaseURL, resource)
	var (
		body            io.Reader
		requestBody, rb []byte
	)
	if reqBody != nil {

		requestBody, err := json.Marshal(reqBody)
		if err != nil {
			return APIAnnotation{}, errors.Wrap(err, "client - unable to marshal request struct")
		}

		// only log request when explicitly asked to do so
		if s.config.LogRequest {
			rb, _ = json.Marshal(reqBody)
			log.Printf("pawapay: making request to route %s with payload %s", URL, rb)
		}

		body = bytes.NewReader(requestBody)
	}

	if reqBody == nil && s.config.LogRequest {
		log.Printf("pawapay: making request to route %s", URL)
	}

	req, err := http.NewRequest(method, URL, body)
	if err != nil {
		return APIAnnotation{}, errors.Wrap(err, "client - unable to create request body")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.APIKey))

	res, err := s.client.Do(req)
	if err != nil {
		return APIAnnotation{}, errors.Wrap(err, "client - failed to execute request")
	}

	b, _ := io.ReadAll(res.Body)
	if s.config.LogResponse {
		log.Printf("pawapay: got response %s code %d", string(b), res.StatusCode)
	}

	var apiAnnotation APIAnnotation
	apiAnnotation.RequestPayload = string(rb)
	apiAnnotation.ResponseCode = res.StatusCode
	apiAnnotation.ResponsePayload = string(b)
	if !strings.EqualFold(os.Getenv("env"), "testing") {
		apiAnnotation.URL = URL
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusCreated {
		if s.config.LogResponse {
			log.Printf("pawapay: error response body: %s for request payload %s", b, requestBody)
		}
		return apiAnnotation, fmt.Errorf("invalid status code received, expected 200/204/201, got %v with body %s", res.StatusCode, b)
	}

	if resp != nil || res.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(bytes.NewReader(b)).Decode(&resp); err != nil {
			return apiAnnotation, errors.Wrap(err, "unable to unmarshal response body")
		}
	}
	return apiAnnotation, nil
}

func (s *Service) newCreatePayoutRequest(timeProvider TimeProviderFunc, payoutId string, amt Amount, countryCode, code, description string,
	pn PhoneNumber) CreatePayoutRequest {
	layout := "2006-01-02T15:04:05Z"
	if len(description) > 22 {
		description = description[:22]
	}

	return CreatePayoutRequest{
		PayoutId:             payoutId,
		Amount:               amt.Value,
		Currency:             amt.Currency,
		Country:              countryCode,
		Correspondent:        code,
		CustomerTimestamp:    timeProvider().Format(layout),
		StatementDescription: description,
		Recipient:            Recipient{Type: recipientType, Address: Address{Value: fmt.Sprintf("%s%s", pn.CountryCode, pn.Number)}},
	}
}

func (s *Service) newCreateBulkPayoutRequest(timeProvider TimeProviderFunc, req []PayoutRequest) ([]CreatePayoutRequest, error) {
	requests := []CreatePayoutRequest{}
	for _, payload := range req {

		query := gountries.New()
		se, err := query.FindCountryByCallingCode(payload.PhoneNumber.CountryCode)
		if err != nil {
			return []CreatePayoutRequest{}, err
		}

		countryCode := se.Alpha3

		requests = append(requests, s.newCreatePayoutRequest(timeProvider, payload.PayoutId, payload.Amount,
			countryCode, payload.Correspondent, payload.Description, payload.PhoneNumber))
	}

	return requests, nil
}

func (s *Service) newDepositRequest(timeProvider TimeProviderFunc, depositId string, amt Amount, countryCode, code, description string,
	pn PhoneNumber, authCode string) CreateDepositRequest {
	layout := "2006-01-02T15:04:05Z"
	if len(description) > 22 {
		description = description[:22]
	}

	return CreateDepositRequest{
		DepositId:            depositId,
		Amount:               amt.Value,
		Currency:             amt.Currency,
		Country:              countryCode,
		Correspondent:        code,
		CustomerTimestamp:    timeProvider().Format(layout),
		StatementDescription: description,
		PreAuthorizationCode: authCode,
		Payer:                Payer{Type: recipientType, Address: Address{Value: fmt.Sprintf("%s%s", pn.CountryCode, pn.Number)}},
	}
}

func (s *Service) newCreateBulkDepositRequest(timeProvider TimeProviderFunc, req []DepositRequest) ([]CreateDepositRequest, error) {
	requests := []CreateDepositRequest{}
	for _, payload := range req {

		query := gountries.New()
		se, err := query.FindCountryByCallingCode(payload.PhoneNumber.CountryCode)
		if err != nil {
			return []CreateDepositRequest{}, err
		}

		countryCode := se.Alpha3

		requests = append(requests, s.newDepositRequest(timeProvider, payload.DepositId, payload.Amount,
			countryCode, payload.Correspondent, payload.Description, payload.PhoneNumber, payload.PreAuthCode))
	}

	return requests, nil
}

type Deposit struct {
	DepositId            string                 `json:"depositId"`
	Status               string                 `json:"status"`
	RequestedAmount      string                 `json:"requestedAmount"`
	DepositedAmount      string                 `json:"depositedAmount"`
	Currency             string                 `json:"currency"`
	Country              string                 `json:"country"`
	Payer                Payer                  `json:"recipient"`
	Correspondent        string                 `json:"correspondent"`
	StatementDescription string                 `json:"statementDescription"`
	CustomerTimestamp    string                 `json:"customerTimestamp"`
	Created              string                 `json:"created"`
	RespondedByPayer     string                 `json:"respondedByPayer"`
	CorrespondentIds     map[string]interface{} `json:"correspondentIds"`
	SuspiciousActivity   map[string]interface{} `json:"suspiciousActivityReport"`
	FailureReason        FailureReason          `json:"failureReason"`
	Annotation           APIAnnotation
}

type DepositStatusResponse struct {
	DepositId  string `json:"depositId"`
	Status     string `json:"status"`
	Annotation APIAnnotation
}

func GetAllCorrespondents() ([]MomoMapping, error) {
	var path string

	if os.Getenv("BANK_FILE_PATH") != "" {
		path = os.Getenv("BANK_FILE_PATH") // USED WHEN TESTING
	}
	fileName := filepath.Join(path, "momo.json")
	bb, err := os.ReadFile(fileName)
	if err != nil {
		return []MomoMapping{}, err
	}

	var momoProviderMappings []MomoMapping
	if err := json.Unmarshal(bb, &momoProviderMappings); err != nil {
		return []MomoMapping{}, err
	}

	return momoProviderMappings, nil
}
