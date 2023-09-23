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
	testPayoutId = "d334c312-6c18-4d7e-a0f1-097d398543d3"
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

		_, err := c.ResendCallback(req)
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
