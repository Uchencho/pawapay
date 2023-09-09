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

type CreatePayoutResponse struct {
	PayoutID   string `json:"payoutId"`
	Status     string `json:"status"`
	Created    string `json:"created"`
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
	layout := "2006-01-02 15:04:04"

	return CreatePayoutRequest{
		PayoutId:             payoutId,
		Amount:               amt.Value,
		Currency:             amt.Currency,
		Country:              countryCode,
		Correspondent:        code,
		CustomerTimestamp:    timeProvider().Format(layout),
		StatementDescription: description[:22],
		Recipient:            Recipient{Type: recipientType, Address: Address{Value: fmt.Sprintf("%s%s", pn.CountryCode, pn.Number)}},
	}
}

func GetMomoMapping(ext string) (MomoMapping, error) {
	var path string

	if os.Getenv("BANK_FILE_PATH") != "" {
		path = os.Getenv("BANK_FILE_PATH") // USED WHEN TESTING
	}
	fileName := filepath.Join(path, "momo.json")
	bb, err := os.ReadFile(fileName)
	if err != nil {
		return MomoMapping{}, err
	}

	var momoProviderMappings []MomoMapping
	if err := json.Unmarshal(bb, &momoProviderMappings); err != nil {
		return MomoMapping{}, err
	}

	for _, mapping := range momoProviderMappings {
		if strings.EqualFold(mapping.Extension, ext) {
			return mapping, nil
		}
	}
	return MomoMapping{}, fmt.Errorf("no mapping found for ext=%s", ext)
}
