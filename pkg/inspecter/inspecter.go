package inspecter

import (
	"net/http"
	"net/url"
	"regexp"
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

	Links []*InspectedLink `json:"links"`

	TotalLinkCount    int `json:"total_link_count"`
	ExternalLinkCount int `json:"external_link_count"`
	InternalLinkCount int `json:"internal_link_count"`
}

type InspectedLink struct {
	URL        string `json:"url"`
	Text       string `json:"text"`
	Type       string `json:"type"`
	StatusCode int    `json:"status_code"`
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

		Links:    []*InspectedLink{},
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

	report.TotalLinkCount = len(report.Links)

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

			case "a":
				linkTk := tkn

				// To get the link text
				tagText, shouldReturn := parseNextTextToken()
				if shouldReturn {
					return
				}

				report.parseLink(&linkTk, tagText)
			}

		case html.ErrorToken:
			return
		}
	}
}

func (report *InspectReport) parseLink(ATag *html.Token, linkText string) {
	link := InspectedLink{Text: linkText, StatusCode: 0}
	var linkURL string

	// Get the href attribute
	for _, attr := range ATag.Attr {
		if attr.Key == "href" {
			linkURL = attr.Val
			break
		}
	}

	// Don't add the link if it's empty
	if len(linkURL) == 0 {
		return
	}

	link.URL = linkURL

	// regex to check if the link is a special action link (javascript, mailto, etc)
	rgxSpecialProtocol := regexp.MustCompile("^([a-zA-Z0-9]*?):")

	shouldAnalyse := false

	if strings.HasPrefix(linkURL, "http") {
		link.Type = "external"
		shouldAnalyse = true
		report.ExternalLinkCount++

	} else if strings.HasPrefix(linkURL, "#") {
		link.Type = "fragment"
		report.InternalLinkCount++

	} else if strings.HasPrefix(linkURL, "tel:") {
		link.Type = "telephone"
		report.ExternalLinkCount++

	} else if strings.HasPrefix(linkURL, "mailto:") {
		link.Type = "email"
		report.ExternalLinkCount++

	} else if specialProtocolMatches := rgxSpecialProtocol.FindAllStringSubmatch(linkURL, 1); len(specialProtocolMatches) > 0 {
		link.Type = specialProtocolMatches[0][1]
		report.ExternalLinkCount++

	} else if strings.HasPrefix(linkURL, "/") {
		link.Type = "absolute"
		report.InternalLinkCount++

		// If the link is absolute, we need to add the domain to the link
		reportURL, reportURLErr := url.Parse(report.URL)
		if reportURLErr == nil {
			linkURL = reportURL.Scheme + "://" + reportURL.Host + linkURL
			shouldAnalyse = true

		} else {
			link.StatusCode = http.StatusBadRequest
			link.Type = "invalid"
		}

	} else {
		link.Type = "relative"
		report.InternalLinkCount++

		shouldAnalyse = true
		// If the link is relative, we need to add the base URL to it
		linkURL = strings.TrimPrefix(report.URL, "/") + "/" + linkURL
	}

	report.Links = append(report.Links, &link)

	if shouldAnalyse {
		go report.analyseLink(linkURL, &link)
	}
}

func (report *InspectReport) analyseLink(inputURL string, link *InspectedLink) {
	//TODO: implement
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
