package pawapay

import (
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

// NewService returns a new pawapay service
func NewService(c Config) Service {
	return Service{
		config: c,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// CreatePayout provides the functionality of creating a payout
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
