package inspecter

import (
	"net/http"
	"strings"

	"golang.org/x/net/html"
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

	tokenizer := html.NewTokenizer(httpResp.Body)
	report.ParseTokens(tokenizer)

	return &report
}

// ParseTokens parses the HTML tokens from the given tokenizer
func (report *InspectReport) ParseTokens(tokenizer *html.Tokenizer) {
	for {
		var tokenType html.TokenType
		var tkn html.Token

		// Defined as a local function to use in the again
		var nextToken = func() {
			tokenType = tokenizer.Next()
			tkn = tokenizer.Token()
		}

		nextToken()

		switch tokenType {
		case html.DoctypeToken:
			report.HTMLVersion = DetectHTMLVersion(tkn.Data)

		case html.StartTagToken:

			tknData := strings.ToLower(tkn.Data)
			switch tknData {

			case "title":
				nextToken() // To get the title text
				if tokenType == html.TextToken {
					report.PageTitle = tkn.Data
				}

			}

		case html.ErrorToken:
			return
		}
	}
}
