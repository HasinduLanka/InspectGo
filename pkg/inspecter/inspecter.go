package inspecter

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/net/html"
)

var MaximumConcurrentLinkAnalysis = 256

// Control the number of concurrent link analysers
var concurrentLinkAnalysersSemaphore chan struct{} = make(chan struct{}, MaximumConcurrentLinkAnalysis)

type InspectReport struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`

	HTMLVersion string `json:"html_version"`
	PageTitle   string `json:"page_title"`

	// format: Headings["h1"] = []string{"big heading", "heading b"}
	Headings map[string][]string `json:"headings"`

	// Number of password fields in the page
	LoginFieldCount int `json:"login_field_count"`

	Links []*InspectedLink `json:"links"`

	AccessibleLinkCount   int `json:"accessible_link_count"`
	InaccessibleLinkCount int `json:"inaccessible_link_count"`
	NotAnalysedLinkCount  int `json:"not_analysed_link_count"`
	TotalLinkCount        int `json:"total_link_count"`
	ExternalLinkCount     int `json:"external_link_count"`
	InternalLinkCount     int `json:"internal_link_count"`

	LinkAnalyticWG       *sync.WaitGroup    `json:"-"`
	RequestContext       *context.Context   `json:"-"`
	RequestContextCancel context.CancelFunc `json:"-"`
}

type InspectedLink struct {
	URL        string `json:"url"`
	Text       string `json:"text"`
	Type       string `json:"type"`
	StatusCode int    `json:"status_code"`
}

// InspectURL returns an InspectReport for the given URL immediately, and continues to analyse the links in the background
//
// linkAnalyticsTimout is the maximum time to wait for the request analytics to complete.
// If the requests takes longer than linkAnalyticsTimout, the analytics are cancelled and the current incomplete report is returned.
//
// Pass nil for linkAnalyticsTimout to avoid link analytics.
func InspectURL(inputURL string, linkAnalyticsTimout *time.Time) *InspectReport {

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

		LinkAnalyticWG: &sync.WaitGroup{},
	}

	if linkAnalyticsTimout != nil {
		// Create a context with a timeout
		reqContextTimout, reqContextCancel := context.WithDeadline(context.Background(), *linkAnalyticsTimout)
		report.RequestContext = &reqContextTimout
		report.RequestContextCancel = reqContextCancel
	} else {
		// Set the context to nil
		report.RequestContext = nil
		report.RequestContextCancel = func() {}
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

		tknData := strings.ToLower(tkn.Data)

		switch tokenType {
		case html.DoctypeToken:
			report.HTMLVersion = DetectHTMLVersion(tknData)

		case html.ErrorToken:
			return

		case html.StartTagToken:

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

			case "input":
				report.parseInputTag(&tkn)

			}

		default:
			switch tknData {

			case "input":
				report.parseInputTag(&tkn)

			}

		}
	}
}

func (report *InspectReport) parseInputTag(tkn *html.Token) {

	// check if a password input
	for _, attr := range tkn.Attr {
		if strings.ToLower(attr.Key) == "type" && strings.ToLower(attr.Val) == "password" {
			report.LoginFieldCount++
			break
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

	// Analyse the link if it's not a special action link
	// and RequestContext is not nil
	if shouldAnalyse && report.RequestContext != nil {
		// Add the link to the wait group
		report.LinkAnalyticWG.Add(1)
		go report.analyseLink(linkURL, &link)
	}
}

func (report *InspectReport) analyseLink(inputURL string, link *InspectedLink) {

	// Remove the link from the wait group
	defer report.LinkAnalyticWG.Done()

	// This blocks if the semaphore is full
	concurrentLinkAnalysersSemaphore <- struct{}{}

	defer func() {
		// Release the semaphore
		<-concurrentLinkAnalysersSemaphore
	}()

	// Get the webpage for the link within the context of the request
	outgoingReq, outgoingReqErr := http.NewRequestWithContext(*report.RequestContext, http.MethodGet, inputURL, nil)

	if outgoingReqErr != nil || outgoingReq == nil {
		link.StatusCode = http.StatusInternalServerError
		link.Type = "error"
		return
	}

	outgoingReq.Header.Set(`User-Agent`, `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36`)
	outgoingReq.Header.Set(`sec-ch-ua`, ` Not A;Brand";v="99", "Chromium";v="100", "Google Chrome";v="100"`)
	outgoingReq.Header.Set(`sec-ch-ua-mobile`, `?0`)
	outgoingReq.Header.Set(`sec-ch-ua-platform`, `"Linux"`)
	outgoingReq.Header.Set(`Sec-Fetch-Dest`, `document`)
	outgoingReq.Header.Set(`Sec-Fetch-Mode`, `navigate`)
	outgoingReq.Header.Set(`Sec-Fetch-Site`, `same-origin`)
	outgoingReq.Header.Set(`Sec-Fetch-User`, `?1`)

	httpResp, httpErr := http.DefaultClient.Do(outgoingReq)

	// If there was an error getting the webpage, return an error
	if httpErr != nil {
		if httpResp != nil {
			link.StatusCode = httpResp.StatusCode
		} else {
			link.StatusCode = http.StatusRequestTimeout
		}

	} else {
		link.StatusCode = httpResp.StatusCode

	}

	// some websites like linkedin, do not allow bots to access their pages
	if link.StatusCode > 600 {
		link.Type = "unscannable"
		link.StatusCode = http.StatusOK
	}

}

func (report *InspectReport) CountLinks() {
	accessible := 0
	inaccessible := 0
	notAnalysed := 0

	for _, lnk := range report.Links {
		if lnk.StatusCode == 0 {
			if lnk.Type == "external" || lnk.Type == "absolute" || lnk.Type == "relative" {
				notAnalysed++
			}
		} else if lnk.StatusCode < 400 {
			accessible++
		} else {
			inaccessible++
		}
	}

	report.AccessibleLinkCount = accessible
	report.InaccessibleLinkCount = inaccessible
	report.NotAnalysedLinkCount = notAnalysed
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
