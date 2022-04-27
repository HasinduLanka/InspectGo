package inspecter

import (
	"net/http"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type InspectReport struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`

	HTMLVersion string `json:"html_version"`
	PageTitle   string `json:"page_title"`

	// format: Headings["h1"] = []string{"big heading", "heading b"}
	Headings map[string][]string `json:"headings"`
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

		Headings: map[string][]string{},
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

		// Some tags like <h1>, <a> tags could contain other tags instead of directly containing text
		// So we need to parse the inner tags by skipping other tags
		var parseNextTextToken = func() (string, bool) {

			// depth is the level of nested tags inside this tag
			depth := 0

			for (tokenType != html.TextToken) && (tokenType != html.ErrorToken) && (depth >= 0) {
				nextToken()

				if tokenType == html.StartTagToken {
					depth++
				} else if tokenType == html.EndTagToken {
					depth--
				}
			}

			if tokenType == html.TextToken {
				tagText := removeHTMLEmptySpace(tkn.Data)
				return tagText, false

			} else if tokenType == html.ErrorToken {
				// If there was an error parsing the HTML return
				// this is most likely due to reaching the end of the HTML
				// it is also probably because the HTML is malformed
				return "", true
			}

			return "", false
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

			case "h1", "h2", "h3", "h4", "h5", "h6":
				headingType := tknData

				// To get the heading text
				tagText, shouldReturn := parseNextTextToken()
				if shouldReturn {
					return
				}

				if len(tagText) > 0 {
					report.Headings[headingType] = append(report.Headings[headingType], tagText)
				}

			}

		case html.ErrorToken:
			return
		}
	}
}

// Remove HTML empty spaces
func removeHTMLEmptySpace(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	previousSpace := false

	for _, ch := range str {
		if ch == ' ' {
			if !previousSpace {
				previousSpace = true
				b.WriteRune(ch)
			}

		} else if unicode.IsSpace(ch) {
			previousSpace = false

		} else {
			previousSpace = false
			b.WriteRune(ch)
		}
	}

	return strings.TrimSpace(b.String())
}
