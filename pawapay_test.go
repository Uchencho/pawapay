package pawapay_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Uchencho/pawapay"
	"github.com/stretchr/testify/assert"
)

const (
	testPayoutId  = "d334c312-6c18-4d7e-a0f1-097d398543d3"
	testDepositId = "d334c312-6c18-4d7e-a0f1-097d398543d3"
)

func timeProvider() pawapay.TimeProviderFunc {
	return func() time.Time {
		t, _ := time.Parse("2006-01-02", "2021-01-01")
		return t
	}
}

func fileToStruct(filepath string, s interface{}) io.Reader {
	bb, _ := os.ReadFile(filepath)
	json.Unmarshal(bb, s)
	return bytes.NewReader(bb)
}

type row struct {
	Name            string
	Input           interface{}
	CustomServerURL func(t *testing.T) string
}

func TestCreatePayout(t *testing.T) {
	table := []row{
		{
			Name: "Creating payout succeeds",
			Input: pawapay.PayoutRequest{
				PayoutId:      testPayoutId,
				Amount:        pawapay.Amount{Currency: "GHS", Value: "1000"},
				Description:   "test",
				PhoneNumber:   pawapay.PhoneNumber{CountryCode: "233", Number: "247492147"},
				Correspondent: "MTN_MOMO_GHA",
			},
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					var actualBody, expectedBody pawapay.CreatePayoutRequest

					if err := json.NewDecoder(req.Body).Decode(&actualBody); err != nil {
						log.Printf("error in unmarshalling %+v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := "/payouts"
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					t.Run("Request is as expected", func(t *testing.T) {
						fileToStruct(filepath.Join("testdata", "create-payout-request.json"), &expectedBody)
						assert.Equal(t, expectedBody, actualBody)
					})

					var resp pawapay.CreatePayoutResponse
					fileToStruct(filepath.Join("testdata", "create-payout-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.(pawapay.PayoutRequest)

		log.Printf("======== Running row: %s ==========", row.Name)

		_, err := c.CreatePayout(timeProvider(), req)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})

	}
}

func TestCreateBulkPayout(t *testing.T) {
	table := []row{
		{
			Name: "Creating bulk payout succeeds",
			Input: []pawapay.PayoutRequest{
				{
					PayoutId:      testPayoutId,
					Amount:        pawapay.Amount{Currency: "GHS", Value: "1000"},
					Description:   "test",
					PhoneNumber:   pawapay.PhoneNumber{CountryCode: "233", Number: "247492147"},
					Correspondent: "MTN_MOMO_GHA",
				},
			},
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					var actualBody, expectedBody []pawapay.CreatePayoutRequest

					if err := json.NewDecoder(req.Body).Decode(&actualBody); err != nil {
						log.Printf("error in unmarshalling %+v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := "/payouts/bulk"
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					t.Run("Request payload is as expected", func(t *testing.T) {
						fileToStruct(filepath.Join("testdata", "create-bulk-payout-request.json"), &expectedBody)
						assert.Equal(t, expectedBody, actualBody)
					})

					var resp []pawapay.CreatePayoutResponse
					fileToStruct(filepath.Join("testdata", "create-bulk-payout-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.([]pawapay.PayoutRequest)

		log.Printf("======== Running row: %s ==========", row.Name)

		_, err := c.CreateBulkPayout(timeProvider(), req)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})

	}
}

func TestGetPayout(t *testing.T) {
	table := []row{
		{
			Name:  "Retrieving payout succeeds",
			Input: testPayoutId,
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := fmt.Sprintf("/payouts/%s", testPayoutId)
						assert.Equal(t, http.MethodGet, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					var resp []pawapay.Payout
					fileToStruct(filepath.Join("testdata", "get-payout-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.(string)

		log.Printf("======== Running row: %s ==========", row.Name)

		_, err := c.GetPayout(req)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})

	}
}

func TestResendPayoutCallBack(t *testing.T) {
	table := []row{
		{
			Name:  "Resend payout callback succeeds",
			Input: testPayoutId,
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					var actualBody, expectedBody pawapay.ResendCallbackRequest
					expectedBody.PayoutId = testPayoutId

					if err := json.NewDecoder(req.Body).Decode(&actualBody); err != nil {
						log.Printf("error in unmarshalling %+v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := "/payouts/resend-callback"
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					t.Run("Request payload is as expected", func(t *testing.T) {
						assert.Equal(t, expectedBody, actualBody)
					})

					var resp pawapay.PayoutStatusResponse
					fileToStruct(filepath.Join("testdata", "payout-status-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.(string)

		log.Printf("======== Running row: %s ==========", row.Name)

		_, err := c.ResendPayoutCallback(req)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})

	}
}

func TestFailEnqueuedPayout(t *testing.T) {
	table := []row{
		{
			Name:  "Fail enqueued payout",
			Input: testPayoutId,
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := fmt.Sprintf("/payouts/fail-enqueued/%s", testPayoutId)
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					var resp pawapay.PayoutStatusResponse
					fileToStruct(filepath.Join("testdata", "payout-status-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.(string)

		log.Printf("======== Running row: %s ==========", row.Name)

		_, err := c.FailEnqueued(req)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})
	}
}

func TestCreateDeposit(t *testing.T) {
	table := []row{
		{
			Name: "deposit is initialized successfully",
			Input: pawapay.DepositRequest{
				DepositId:     testDepositId,
				Amount:        pawapay.Amount{Currency: "GHS", Value: "1000"},
				Description:   "test",
				PhoneNumber:   pawapay.PhoneNumber{CountryCode: "233", Number: "247492147"},
				Correspondent: "MTN_MOMO_GHA",
			},
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					var actualBody, expectedBody pawapay.CreateDepositRequest

					if err := json.NewDecoder(req.Body).Decode(&actualBody); err != nil {
						log.Printf("error in unmarshalling %+v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := "/deposits"
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					t.Run("Request is as expected", func(t *testing.T) {
						fileToStruct(filepath.Join("testdata", "create-deposit-request.json"), &expectedBody)
						assert.Equal(t, expectedBody, actualBody)
					})

					var resp pawapay.CreateDepositResponse
					fileToStruct(filepath.Join("testdata", "create-deposit-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.(pawapay.DepositRequest)

		log.Printf("======== Running row: %s ==========", row.Name)

		_, err := c.InitiateDeposit(timeProvider(), req)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})

	}
}

func TestCreateBulkDeposit(t *testing.T) {
	table := []row{
		{
			Name: "bulk deposit is initialized successfully",
			Input: []pawapay.DepositRequest{
				{
					DepositId:     testDepositId,
					Amount:        pawapay.Amount{Currency: "GHS", Value: "1000"},
					Description:   "test",
					PhoneNumber:   pawapay.PhoneNumber{CountryCode: "233", Number: "247492147"},
					Correspondent: "MTN_MOMO_GHA",
					PreAuthCode:   "QJS3RSK",
				},
			},
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					var actualBody, expectedBody []pawapay.CreateDepositRequest

					if err := json.NewDecoder(req.Body).Decode(&actualBody); err != nil {
						log.Printf("error in unmarshalling %+v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := "/deposits/bulk"
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					t.Run("Request is as expected", func(t *testing.T) {
						fileToStruct(filepath.Join("testdata", "create-bulk-deposit-request.json"), &expectedBody)
						assert.Equal(t, expectedBody, actualBody)
					})

					var resp []pawapay.CreateDepositResponse
					fileToStruct(filepath.Join("testdata", "create-bulk-deposit-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.([]pawapay.DepositRequest)

		log.Printf("======== Running row: %s ==========", row.Name)

		_, err := c.InitiateBulkDeposit(timeProvider(), req)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})
	}
}

func TestGetDeposit(t *testing.T) {
	table := []row{
		{
			Name:  "Retrieving deposit succeeds",
			Input: testDepositId,
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := fmt.Sprintf("/deposits/%s", testDepositId)
						assert.Equal(t, http.MethodGet, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					var resp []pawapay.Deposit
					fileToStruct(filepath.Join("testdata", "get-deposit-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.(string)

		log.Printf("======== Running row: %s ==========", row.Name)

		_, err := c.GetDeposit(req)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})
	}
}

func TestResendDepositCallBack(t *testing.T) {
	table := []row{
		{
			Name:  "Resend payout callback succeeds",
			Input: testDepositId,
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					var actualBody, expectedBody pawapay.ResendCallbackRequest
					expectedBody.DepositId = testDepositId

					if err := json.NewDecoder(req.Body).Decode(&actualBody); err != nil {
						log.Printf("error in unmarshalling %+v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := "/deposits/resend-callback"
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					t.Run("Request payload is as expected", func(t *testing.T) {
						assert.Equal(t, expectedBody, actualBody)
					})

					var resp pawapay.DepositStatusResponse
					fileToStruct(filepath.Join("testdata", "deposit-status-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.(string)

		log.Printf("======== Running row: %s ==========", row.Name)

		_, err := c.ResendDepositCallback(req)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})

	}
}

func TestRequestRefund(t *testing.T) {

	type refundPayload struct {
		DepositId string
		Amount    pawapay.Amount
		RefundId  string
	}

	table := []row{
		{
			Name: "refund is requested successfully",
			Input: refundPayload{
				DepositId: testDepositId,
				Amount:    pawapay.Amount{Currency: "GHS", Value: "1000"},
				RefundId:  testDepositId,
			},
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					var actualBody, expectedBody pawapay.RefundRequest

					if err := json.NewDecoder(req.Body).Decode(&actualBody); err != nil {
						log.Printf("error in unmarshalling %+v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := "/refunds"
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					t.Run("Request is as expected", func(t *testing.T) {
						fileToStruct(filepath.Join("testdata", "refund-request.json"), &expectedBody)
						assert.Equal(t, expectedBody, actualBody)
					})

					var resp pawapay.InitiateRefundResponse
					fileToStruct(filepath.Join("testdata", "refund-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.(refundPayload)

		log.Printf("======== Running row: %s ==========", row.Name)

		_, err := c.RequestRefund(req.RefundId, req.DepositId, req.Amount)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})

	}
}

func TestGetRefund(t *testing.T) {
	table := []row{
		{
			Name:  "Retrieving refund succeeds",
			Input: testDepositId,
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := fmt.Sprintf("/refunds/%s", testDepositId)
						assert.Equal(t, http.MethodGet, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					var resp []pawapay.Refund
					fileToStruct(filepath.Join("testdata", "get-refund-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.(string)

		log.Printf("======== Running row: %s ==========", row.Name)

		result, err := c.GetRefund(req)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})

		t.Run("Annotation response code is as expected", func(t *testing.T) {
			assert.Equal(t, http.StatusOK, result.Annotation.ResponseCode)
		})
	}
}

func TestResendRefundCallBack(t *testing.T) {
	table := []row{
		{
			Name:  "Resend refund callback succeeds",
			Input: testDepositId,
			CustomServerURL: func(t *testing.T) string {
				pawapayService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

					var actualBody, expectedBody pawapay.ResendCallbackRequest
					expectedBody.RefundId = testDepositId

					if err := json.NewDecoder(req.Body).Decode(&actualBody); err != nil {
						log.Printf("error in unmarshalling %+v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					t.Run("URL and request method is as expected", func(t *testing.T) {
						expectedURL := "/refunds/resend-callback"
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, expectedURL, req.RequestURI)
					})

					t.Run("Request payload is as expected", func(t *testing.T) {
						assert.Equal(t, expectedBody, actualBody)
					})

					var resp pawapay.RefundStatusResponse
					fileToStruct(filepath.Join("testdata", "refund-status-response.json"), &resp)

					w.WriteHeader(http.StatusOK)
					bb, _ := json.Marshal(resp)
					w.Write(bb)

				}))
				return pawapayService.URL
			},
		},
	}

	for _, row := range table {

		c := pawapay.NewService(pawapay.Config{
			BaseURL: row.CustomServerURL(t),
		})

		req := row.Input.(string)

		log.Printf("======== Running row: %s ==========", row.Name)

		result, err := c.ResendRefundCallback(req)
		t.Run("No error is returned", func(t *testing.T) {
			assert.NoError(t, err)
		})
		t.Run("Annotation response code is as expected", func(t *testing.T) {
			assert.Equal(t, http.StatusOK, result.Annotation.ResponseCode)
		})
	}
}
