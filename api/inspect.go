package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type inspectEndpointRequest struct {
	URL string `json:"url"`
}

func InspectEndpoint(wr http.ResponseWriter, req *http.Request) {

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

	inspectResp := "InspectEndpoint: " + reqBody.URL

	respBody, respEncodeErr := json.Marshal(inspectResp)

	// If there was an error encoding the response body, return an error
	if respEncodeErr != nil {
		log.Println("endpoint /inspect : response encode error : " + respEncodeErr.Error())
		http.Error(wr, respEncodeErr.Error(), http.StatusInternalServerError)
		return
	}

	wr.Write([]byte(respBody))
}
