# Pawapay

[Pawapay](https://pawapay.io/) client application written in Go


![alt text](https://global-uploads.webflow.com/62824591015aa314fd308df1/6411b26596e3de3f52551c00_Logopawapay-p-500.png)



## Usage

### Install Package

```bash
go get github.com/Uchencho/pawapay
```

### Documentation

Please see [the docs](https://pawapay.io/) for the most up-to-date documentation of the Pawapay API.

#### Pawapay

- Sample Usage

```go
package main

import (
	"log"
	"time"

	"github.com/Uchencho/pawapay"
)

func main() {
	cfg := pawapay.GetConfigFromEnvVars()
	cfg.AllowRequestLogging() // only for debugging, would not advise this for production.
	// Also not absolutely necessary because we store API annotations in the response returned

	/*
		You can also explicitly declare the config
		cfg := pawapay.Config{
				APIKey:      "key",
				BaseURL:     "url",
				LogRequest:  os.Getenv("env") == "production",
				LogResponse: strings.EqualFold(os.Getenv("env"), "production"),
			}
	*/

	service := pawapay.NewService(cfg)

	amt := pawapay.Amount{Currency: "GHS", Value: "500"}
	description := "sending money to all my children" // this will be truncated to the first 22 characters
	pn := pawapay.PhoneNumber{CountryCode: "233", Number: "704584739348"}

	allCorrespondentMappings, err := pawapay.GetAllCorrespondents()
	if err != nil {
		log.Fatal(err)
	}

	// each mapping, think country, have a number of correspondents, pick the one you are trying to send money to.
	// Ideally you will take this as input and map to the correspondent of your choice
	correspondent := allCorrespondentMappings[0].Correspondents[0]
	req := pawapay.PayoutRequest{
		Amount:        amt,
		PhoneNumber:   pn,
		Description:   description,
		PayoutId:      "uniqueId",
		Correspondent: correspondent.Correspondent,
	}

	resp, err := service.CreatePayout(time.Now, req)
	if err != nil {
		log.Printf("something went wrong, we will confirm through their webhook")

		// even in error, depending on the error, you might have access to the annotation

		log.Printf("request failed with status code %v, response payload %s, error=%s",
			resp.Annotation.ResponseCode, resp.Annotation.ResponsePayload, err)
	}

	log.Printf("response: %+v", resp)
}


```

> **NOTE**

> You also have access to deposit and refund functionalities

> Check the `client` directory to see a sample implementation and pawapay_test.go file to see sample tests
