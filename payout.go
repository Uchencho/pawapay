package pawapay

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/biter777/countries"
	"github.com/pkg/errors"
)

// Config represents the pawapay config
type Config struct {
	BaseURL     string
	APIKey      string
	LogRequest  bool
	LogResponse bool
}

// Service is a representation of a pawapay service
type Service struct {
	config Config
	client *http.Client
}

// ConfigProvider pawapay config provider
type ConfigProvider func() Config

// GetConfigFromEnvVars returns config configurations from environment variables
func GetConfigFromEnvVars() Config {
	return Config{
		BaseURL: os.Getenv("PAWAPAY_API_URL"),
		APIKey:  os.Getenv("PAWAPAY_API_KEY"),
	}
}

// functionality to allow you log request payload. This is only necessary for debugging as all response
// objects return the api annotation giving all the required details of the request
func (c *Config) AllowRequestLogging() { c.LogRequest = true }

// functionality to allow you log response payload. This is only necessary for debugging as all response
// objects return the api annotation giving all the required details of the request
func (c *Config) AllowResponseLogging() { c.LogResponse = true }

// functionality to allow the package log request and responses
func (c *Config) AllowLogging() {
	c.AllowRequestLogging()
	c.AllowResponseLogging()
}

// NewService returns a new pawapay service
func NewService(c Config) Service {
	return Service{
		config: c,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// CreatePayout provides the functionality of creating a payout
// See docs https://docs.pawapay.co.uk/#operation/createPayout for more details
func (s *Service) CreatePayout(timeProvider TimeProviderFunc, payoutId string, amt Amount,
	description string, pn PhoneNumber, correspondent string) (CreatePayoutResponse, error) {

	cc, err := strconv.Atoi(pn.CountryCode)
	if err != nil {
		return CreatePayoutResponse{}, errors.Wrapf(err, "unable to convert countryCode=%s to integer", pn.CountryCode)
	}
	c := countries.ByNumeric(cc)
	countryCode := c.Alpha3()

	resource := "payouts"
	payload := s.newCreatePayoutRequest(timeProvider, payoutId, amt, countryCode, correspondent, description, pn)

	var response CreatePayoutResponse
	annotation, err := s.makeRequest(http.MethodPost, resource, payload, &response)
	if err != nil {
		return CreatePayoutResponse{}, err
	}
	response.Annotation = annotation

	return response, nil
}

// CreateBulkPayout provides the functionality of creating a bulk payout
// See docs https://docs.pawapay.co.uk/#operation/createPayout for more details
func (s *Service) CreateBulkPayout(timeProvider TimeProviderFunc, data []PayoutRequest) (CreateBulkPayoutResponse, error) {

	resource := "payouts/bulk"
	payload, err := s.newCreateBulkPayoutRequest(timeProvider, data)
	if err != nil {
		return CreateBulkPayoutResponse{}, err
	}

	var response []CreatePayoutResponse
	annotation, err := s.makeRequest(http.MethodPost, resource, payload, &response)
	if err != nil {
		return CreateBulkPayoutResponse{}, err
	}

	return CreateBulkPayoutResponse{Result: response, Annotation: annotation}, nil
}

// GetPayout provides the functionality of retrieving a payout
// See docs https://docs.pawapay.co.uk/#operation/getPayout for more details
func (s *Service) GetPayout(payoutId string) (Payout, error) {

	resource := fmt.Sprintf("payouts/%s", payoutId)
	var (
		response []Payout
		result   Payout
	)

	annotation, err := s.makeRequest(http.MethodGet, resource, nil, &response)
	if err != nil {
		return Payout{}, err
	}
	if len(response) > 0 {
		result = response[0]
	}
	result.Annotation = annotation
	return result, nil
}

// ResendCallback provides the functionality of resending a callback (webhook)
// See docs https://docs.pawapay.co.uk/#operation/payoutsResendCallback for more details
func (s *Service) ResendCallback(payoutId string) (PayoutStatusResponse, error) {

	resource := "payouts/resend-callback"
	payload := ResendCallbackRequest{PayoutId: payoutId}

	var response PayoutStatusResponse
	annotation, err := s.makeRequest(http.MethodPost, resource, payload, &response)
	if err != nil {
		return PayoutStatusResponse{}, err
	}
	response.Annotation = annotation

	return response, nil
}

// FailEnqueued provides the functionality of failing an already created payout
// See docs https://docs.pawapay.co.uk/#operation/payoutsFailEnqueued for more details
func (s *Service) FailEnqueued(payoutId string) (PayoutStatusResponse, error) {

	resource := fmt.Sprintf("payouts/fail-enqueued/%s", payoutId)

	var response PayoutStatusResponse
	annotation, err := s.makeRequest(http.MethodPost, resource, nil, &response)
	if err != nil {
		return PayoutStatusResponse{}, err
	}
	response.Annotation = annotation

	return response, nil
}
