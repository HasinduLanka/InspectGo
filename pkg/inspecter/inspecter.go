package inspecter

import (
	"net/http"
	"strings"
)

type InspectReport struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`

	HTMLVersion string `json:"html_version"`
	PageTitle   string `json:"page_title"`
}

// InspectURL returns an InspectReport for the given URL
func InspectURL(inputURL string) *InspectReport {

	if !strings.HasPrefix(inputURL, "https://") && !strings.HasPrefix(inputURL, "http://") {
		inputURL = "https://" + inputURL
	}

	// Initialize the report with default values
	report := InspectReport{
		URL: inputURL,

		HTMLVersion: `Not defined`,
		PageTitle:   `Not defined`,
	}

	// Get the webpage
	httpResp, httpErr := http.Get(inputURL)

	// If there was an error getting the webpage, return an error
	if httpErr != nil {

		if httpResp != nil {
			report.StatusCode = httpResp.StatusCode
			report.StatusMsg = httpResp.Status
		} else {
			report.StatusCode = http.StatusBadRequest
			report.StatusMsg = httpErr.Error()
		}
		return &report
	}

	report.StatusCode = httpResp.StatusCode
	report.StatusMsg = httpResp.Status

	return &report
}
