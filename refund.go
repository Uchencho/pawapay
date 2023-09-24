package pawapay

import (
	"fmt"
	"net/http"
)

// RequestRefund provides the functionality of requesting a refund for an initiated deposit
// See docs https://docs.pawapay.co.uk/#operation/depositWebhook for more details
func (s *Service) RequestRefund(refundId, depositId string, amount Amount) (InitiateRefundResponse, error) {

	resource := "refunds"
	payload := s.newRefundRequest(refundId, depositId, amount)

	var response InitiateRefundResponse
	annotation, err := s.makeRequest(http.MethodPost, resource, payload, &response)
	if err != nil {
		return InitiateRefundResponse{}, err
	}
	response.Annotation = annotation

	return response, nil
}

// GetRefund provides the functionality of retrieving an initiated refund
// See docs https://docs.pawapay.co.uk/#operation/getRefund for more details
func (s *Service) GetRefund(refundId string) (Refund, error) {

	resource := fmt.Sprintf("refunds/%s", refundId)
	var (
		response []Refund
		result   Refund
	)

	annotation, err := s.makeRequest(http.MethodGet, resource, nil, &response)
	if err != nil {
		return Refund{}, err
	}
	if len(response) > 0 {
		result = response[0]
	}
	result.Annotation = annotation
	return result, nil
}

// ResendRefundCallback provides the functionality of resending a callback (webhook)
// See docs https://docs.pawapay.co.uk/#operation/refundsResendCallback for more details
func (s *Service) ResendRefundCallback(refundId string) (RefundStatusResponse, error) {

	resource := "refunds/resend-callback"
	payload := ResendCallbackRequest{RefundId: refundId}

	var response RefundStatusResponse
	annotation, err := s.makeRequest(http.MethodPost, resource, payload, &response)
	if err != nil {
		return RefundStatusResponse{}, err
	}
	response.Annotation = annotation

	return response, nil
}
