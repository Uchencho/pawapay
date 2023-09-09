package main

import (
	"log"
	"time"

	"github.com/Uchencho/pawapay"
)

func main() {
	cfg := pawapay.GetConfigFromEnvVars()
	cfg.AllowRequestLogging() // only for debugging, would not advise this for production. Also not necessary, read on for why
	service := pawapay.NewService(cfg)

	amt := pawapay.Amount{Currency: "GHS", Value: "500"}
	description := "sending money to all my children" // this will be truncated to the first 22 characters
	pn := pawapay.PhoneNumber{CountryCode: "233", Number: "704584739348"}

	mapping, err := pawapay.GetMomoMapping(pn.CountryCode)
	if err != nil {
		log.Fatal(err)
	}

	// each mapping, think country, have a number of correspondents, pick the one you are trying to send money to
	// ideally you will take this as an input and map to the correspondent of your choice
	correspondent := mapping.Correspondents[2]

	resp, err := service.CreatePayout(time.Now, "uniqueId", amt, description, pn, correspondent.Correspondent)
	if err != nil {
		log.Printf("something went wrong, we will confirm through their webhook")

		// PLEASE DON'T BELIEVE THAT THEY DID NOT PROCESS THE TRANSFER, CONFIRM

		// even in error, depending on the error, you might have access to the annotation

		log.Printf("request failed with status code %v, response payload %s, error=%s",
			resp.Annotation.ResponseCode, resp.Annotation.ResponsePayload, err)
	}

	log.Printf("response: %+v", resp)
}
