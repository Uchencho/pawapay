package pawapay

import (
	"fmt"
	"net/http"

	"github.com/pariz/gountries"
)

// InitiateDeposit provides the functionality of initiating a deposit for the sender to confirm
// See docs https://docs.pawapay.co.uk/#operation/createDesposit for more details
func (s *Service) InitiateDeposit(timeProvider TimeProviderFunc, depositReq DepositRequest) (CreateDepositResponse, error) {

	query := gountries.New()
	se, err := query.FindCountryByCallingCode(depositReq.PhoneNumber.CountryCode)
	if err != nil {
		return CreateDepositResponse{}, err
	}
	countryCode := se.Alpha3

	resource := "deposits"
	payload := s.newDepositRequest(timeProvider, depositReq.DepositId, depositReq.Amount, countryCode,
		depositReq.Correspondent, depositReq.Description, depositReq.PhoneNumber, depositReq.PreAuthCode)

	var response CreateDepositResponse
	annotation, err := s.makeRequest(http.MethodPost, resource, payload, &response)
	if err != nil {
		return CreateDepositResponse{}, err
	}
	response.Annotation = annotation

	return response, nil
}

// CreateBulkDeposit provides the functionality of creating a bulk deposit
// See docs https://docs.pawapay.co.uk/#operation/createDeposits for more details
func (s *Service) InitiateBulkDeposit(timeProvider TimeProviderFunc, data []DepositRequest) (CreateBulkDepositResponse, error) {

	resource := "deposits/bulk"
	payload, err := s.newCreateBulkDepositRequest(timeProvider, data)
	if err != nil {
		return CreateBulkDepositResponse{}, err
	}

	var response []CreateDepositResponse
	annotation, err := s.makeRequest(http.MethodPost, resource, payload, &response)
	if err != nil {
		return CreateBulkDepositResponse{}, err
	}

	return CreateBulkDepositResponse{Result: response, Annotation: annotation}, nil
}

// GetDeposit provides the functionality of retrieving a deposit
// See docs https://docs.pawapay.co.uk/#operation/getDeposit for more details
func (s *Service) GetDeposit(depositId string) (Deposit, error) {

	resource := fmt.Sprintf("deposits/%s", depositId)
	var (
		response []Deposit
		result   Deposit
	)

	annotation, err := s.makeRequest(http.MethodGet, resource, nil, &response)
	if err != nil {
		return Deposit{}, err
	}
	if len(response) > 0 {
		result = response[0]
	}
	result.Annotation = annotation
	return result, nil
}

// ResendDepositCallback provides the functionality of resending a callback (webhook) for a deposit
// See docs https://docs.pawapay.co.uk/#operation/depositsResendCallback for more details
func (s *Service) ResendDepositCallback(depositId string) (DepositStatusResponse, error) {

	resource := "deposits/resend-callback"
	payload := ResendCallbackRequest{DepositId: depositId}

	var response DepositStatusResponse
	annotation, err := s.makeRequest(http.MethodPost, resource, payload, &response)
	if err != nil {
		return DepositStatusResponse{}, err
	}
	response.Annotation = annotation

	return response, nil
}
