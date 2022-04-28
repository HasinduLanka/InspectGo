package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/HasinduLanka/InspectGo/pkg/inspector"
)

type inspectEndpointRequest struct {
	URL string `json:"url"`
}

func InspectEndpoint(wr http.ResponseWriter, req *http.Request) {

	bestBefore := time.Now().Add(MaxAPIRequestDuration)

	var reqBody inspectEndpointRequest

	// Decode the request body into `inspectEndpointRequest`
	decoder := json.NewDecoder(req.Body)
	decodeErr := decoder.Decode(&reqBody)

	// If there was an error decoding the request body, return an error
	if decodeErr != nil {
		log.Println("endpoint /inspect : request parse error : " + decodeErr.Error())
		http.Error(wr, decodeErr.Error(), http.StatusBadRequest)
		return
	}

	inspectResp := inspector.InspectURL(reqBody.URL, &bestBefore)
	// If there was an error inspecting the URL, it will be returned in the response

	// Return the response in chunks (response streaming)
	flusher, flusherAvailable := wr.(http.Flusher)

	if !flusherAvailable {
		log.Println("endpoint /inspect : response streaming unavailable in this platform")
	}

	// check request header for "inspector-response-streamable"
	if req.Header.Get("inspector-response-streamable") != "true" {
		flusherAvailable = false
	}

	respondReport := func() {

		inspectResp.CountLinks()
		respBody, respEncodeErr := json.Marshal(inspectResp)

		// If there was an error encoding the response body, return an error
		if respEncodeErr != nil {
			log.Println("endpoint /inspect : response encode error : " + respEncodeErr.Error())
			http.Error(wr, respEncodeErr.Error(), http.StatusInternalServerError)
			return
		}

		wr.Write([]byte(respBody))
		if flusherAvailable {
			flusher.Flush()
		}
	}

	var endChannel chan bool

	if flusherAvailable {
		// Return the initial report. This won't contain link analysis information
		respondReport()
		log.Println("endpoint /inspect : initial report returned")

		time.Sleep(2 * time.Second)

		// Ticker to respond every 30 seconds
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		endChannel := make(chan bool)

		go func() {
			for {
				select {
				case <-ticker.C:
					respondReport()
					log.Println("endpoint /inspect : ticking report returned")

				case <-endChannel:
					log.Println("endpoint /inspect : request context done")
					return
				}
			}
		}()
	}

	// Wait for the link analysis to finish
	if inspectResp.LinkAnalyticWG != nil {
		inspectResp.LinkAnalyticWG.Wait()
	}

	if endChannel != nil {
		endChannel <- true
	}

	// Return the final report
	respondReport()
	log.Println("endpoint /inspect : final report returned")

	if inspectResp.RequestContextCancel != nil {
		inspectResp.RequestContextCancel()
	}
}

var MaxAPIRequestDuration = getMaxAPIRequestDuration()

func getMaxAPIRequestDuration() time.Duration {
	_, isVercel := os.LookupEnv(`VERCEL`)
	if isVercel {
		return time.Second * 9
	} else {
		return time.Minute * 10
	}
}
