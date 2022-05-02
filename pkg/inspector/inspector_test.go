package inspector

import (
	"log"
	"net/http"
	"testing"
)

const testSiteURL = "https://inspect-go.vercel.app/tests/"

// Web pages change over time. So we use archived web pages from our server.
func testInspectURL(urlPair urlPair) *InspectReport {

	log.Println("Testing URL :: archived : ", urlPair.archived, " | original : ", urlPair.original)

	if urlPair.original == "" {
		urlPair.original = urlPair.archived
	}

	// Get the webpage
	httpResp, httpErr := http.Get(urlPair.archived)

	// Return the report
	return InspectURLResponse(urlPair.original, httpResp, httpErr, nil)
}

type urlPair struct {
	archived string
	original string
}

func TestInspectURLStatusCode(t *testing.T) {
	expectedStatusCodes := map[urlPair]int{
		{"https://www.google.com", ""}:                200,
		{"https://go.dev", ""}:                        200,
		{"https://en.wikipedia.org/wiki/Germany", ""}: 200,

		{"https://thissiteshouldnot-exist.com/", ""}: 400,
		{"this-siteshouldnot-exist.io", ""}:          400,
		{"httptricksite.dev", ""}:                    400,
		{"nosuchhost", ""}:                           400,

		{"https://en.wikipedia.org/wiki/This-article-should-not-exist-847483", ""}: 404,
		{"https://go.dev/unknownpage", ""}:                                         404,
	}

	for url, expectedStatusCode := range expectedStatusCodes {
		report := testInspectURL(url)
		if report.StatusCode != expectedStatusCode {
			t.Errorf("URL %s returned status code %d, expected %d", url, report.StatusCode, expectedStatusCode)
		} else {
			log.Println("URL", url, "returned expected status code")
		}
	}
}

func TestInspectURLTitle(t *testing.T) {
	expectedTitles := map[urlPair]string{
		{testSiteURL + "go.dev.html", "https://go.dev"}:                                           "The Go Programming Language",
		{testSiteURL + "germany.wiki.html", "https://en.wikipedia.org/wiki/Germany"}:              "Germany - Wikipedia",
		{testSiteURL + "go.wiki.html", "https://en.wikipedia.org/wiki/Go_(programming_language)"}: "Go (programming language) - Wikipedia",
		{"https://www.w3.org/TR/html401", ""}:                                                     "HTML 4.01 Specification",
	}

	for url, expectedTitle := range expectedTitles {
		report := testInspectURL(url)
		if report.PageTitle != expectedTitle {
			t.Errorf("URL %s returned title %s, expected %s", url, report.PageTitle, expectedTitle)
		} else {
			log.Println("URL", url, "returned expected title")
		}
	}
}

func TestInspectURLDocType(t *testing.T) {
	expectedDocTypes := map[urlPair]string{
		{testSiteURL + "go.dev.html", "https://go.dev"}:                                           "HTML 5",
		{testSiteURL + "germany.wiki.html", "https://en.wikipedia.org/wiki/Germany"}:              "HTML 5",
		{testSiteURL + "go.wiki.html", "https://en.wikipedia.org/wiki/Go_(programming_language)"}: "HTML 5",
		{testSiteURL + "w3.html401.html", "https://www.w3.org/TR/html401"}:                        "HTML 4.01 Transitional",
	}

	for url, expectedDocType := range expectedDocTypes {
		report := testInspectURL(url)
		if report.HTMLVersion != expectedDocType {
			t.Errorf("URL %s returned doc type %s, expected %s", url, report.HTMLVersion, expectedDocType)
		} else {
			log.Println("URL", url, "returned expected doc type.")
		}
	}
}

func TestInspectURLHeadings(t *testing.T) {

	// Use web pages from archive.org, so they will not change with time
	expectedHeadingCounts := map[urlPair]map[string]int{
		{testSiteURL + "germany.wiki.html", "https://en.wikipedia.org/wiki/Germany"}: {
			"h1": 1,
			"h2": 14,
			"h3": 24,
		},
		{testSiteURL + "go.wiki.html", "https://en.wikipedia.org/wiki/Go_(programming_language)"}: {
			"h1": 1,
			"h2": 14,
			"h3": 14,
			"h4": 4,
		},
	}

	for url, expectedHeadingCount := range expectedHeadingCounts {
		report := testInspectURL(url)
		for heading, expectedCount := range expectedHeadingCount {
			if len(report.Headings[heading]) != expectedCount {
				t.Errorf("URL %s returned %d headings of type %s, expected %d", url, len(report.Headings[heading]), heading, expectedCount)
			}
		}
	}

}

func TestInspectURLLoginFeilds(t *testing.T) {

	// Use web pages from archive.org, so they will not change with time
	expectedLoginFieldCount := map[urlPair]int{
		{testSiteURL + "log.in.wikipedia.html", "https://en.wikipedia.org/w/index.php?title=Special:UserLogin"}:                                 1,
		{testSiteURL + "create.account.wikipedia.html", "https://en.wikipedia.org/w/index.php?title=Special%3ACreateAccount&campaign=loginCTA"}: 2,
		{"https://en.wikipedia.org/wiki/Main_Page", ""}:             0,
		{testSiteURL + "dev.to.enter.html", "https://dev.to/enter"}: 1,
	}

	for url, expectedCount := range expectedLoginFieldCount {
		report := testInspectURL(url)
		if report.LoginFieldCount != expectedCount {
			t.Errorf("URL %s returned %d login fields, expected %d", url, report.LoginFieldCount, expectedCount)
		}
	}

}

func TestInspectURLLinks(t *testing.T) {

	// Use web pages from archive.org, so they will not change with time
	expectedLinkCounts := map[urlPair]InspectReport{
		{testSiteURL + "germany.wiki.html", "https://en.wikipedia.org/wiki/Germany"}: {
			TotalLinkCount:    3694,
			ExternalLinkCount: 903,
			InternalLinkCount: 2791,
		},
		{testSiteURL + "go.wiki.html", "https://en.wikipedia.org/wiki/Go_(programming_language)"}: {
			TotalLinkCount:    1220,
			ExternalLinkCount: 258,
			InternalLinkCount: 962,
		},
	}

	for url, expectedReport := range expectedLinkCounts {
		report := testInspectURL(url)
		if report.TotalLinkCount != expectedReport.TotalLinkCount {
			t.Errorf("URL %s returned %d total links, expected %d", url, report.TotalLinkCount, expectedReport.TotalLinkCount)
		}
		if report.ExternalLinkCount != expectedReport.ExternalLinkCount {
			t.Errorf("URL %s returned %d external links, expected %d", url, report.ExternalLinkCount, expectedReport.ExternalLinkCount)
		}
		if report.InternalLinkCount != expectedReport.InternalLinkCount {
			t.Errorf("URL %s returned %d internal links, expected %d", url, report.InternalLinkCount, expectedReport.InternalLinkCount)
		}
	}
}
