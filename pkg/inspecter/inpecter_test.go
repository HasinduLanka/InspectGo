package inspecter

import (
	"log"
	"testing"
)

func TestInspectURLStatusCode(t *testing.T) {
	expectedStatusCodes := map[string]int{
		"https://www.google.com":                200,
		"google.com":                            200,
		"https://go.dev":                        200,
		"https://en.wikipedia.org/wiki/Germany": 200,

		"https://thissiteshouldnot-exist.com/": 400,
		"this-siteshouldnot-exist.io":          400,
		"httptricksite.dev":                    400,
		"nosuchhost":                           400,

		"https://en.wikipedia.org/wiki/This-article-should-not-exist-847483": 404,
		"https://go.dev/unknownpage":                                         404,
	}

	for url, expectedStatusCode := range expectedStatusCodes {
		report := InspectURL(url)
		if report.StatusCode != expectedStatusCode {
			t.Errorf("URL %s returned status code %d, expected %d", url, report.StatusCode, expectedStatusCode)
		} else {
			log.Println("URL", url, "returned expected status code")
		}
	}
}

func TestInspectURLTitle(t *testing.T) {
	expectedTitles := map[string]string{
		"https://go.dev":                                          "The Go Programming Language",
		"https://en.wikipedia.org/wiki/Germany":                   "Germany - Wikipedia",
		"https://en.wikipedia.org/wiki/Go_(programming_language)": "Go (programming language) - Wikipedia",
		"https://www.w3.org/TR/html401":                           "HTML 4.01 Specification",
	}

	for url, expectedTitle := range expectedTitles {
		report := InspectURL(url)
		if report.PageTitle != expectedTitle {
			t.Errorf("URL %s returned title %s, expected %s", url, report.PageTitle, expectedTitle)
		} else {
			log.Println("URL", url, "returned expected title")
		}
	}
}

func TestInspectURLDocType(t *testing.T) {
	expectedDocTypes := map[string]string{
		"https://go.dev":                                          "HTML 5",
		"https://en.wikipedia.org/wiki/Germany":                   "HTML 5",
		"https://en.wikipedia.org/wiki/Go_(programming_language)": "HTML 5",
		"https://www.w3.org/TR/html401/":                          "HTML 4.01 Transitional",
	}

	for url, expectedDocType := range expectedDocTypes {
		report := InspectURL(url)
		if report.HTMLVersion != expectedDocType {
			t.Errorf("URL %s returned doc type %s, expected %s", url, report.HTMLVersion, expectedDocType)
		} else {
			log.Println("URL", url, "returned expected doc type")
		}
	}
}

func TestInspectURLHeadings(t *testing.T) {

	// Use web pages from archive.org, so they will not change with time
	expectedHeadingCounts := map[string]map[string]int{
		"https://web.archive.org/web/20220426124521/https://en.wikipedia.org/wiki/Germany": {
			"h1": 1,
			"h2": 14,
			"h3": 24,
		},
		"https://web.archive.org/web/20220426164538/https://en.wikipedia.org/wiki/Go_(programming_language)": {
			"h1": 1,
			"h2": 14,
			"h3": 14,
			"h4": 4,
		},
	}

	for url, expectedHeadingCount := range expectedHeadingCounts {
		report := InspectURL(url)
		for heading, expectedCount := range expectedHeadingCount {
			if len(report.Headings[heading]) != expectedCount {
				t.Errorf("URL %s returned %d headings of type %s, expected %d", url, len(report.Headings[heading]), heading, expectedCount)
			}
		}
	}

}

func TestInspectURLLinks(t *testing.T) {

	// Use web pages from archive.org, so they will not change with time
	expectedLinkCounts := map[string]InspectReport{
		"https://web.archive.org/web/20220426124521/https://en.wikipedia.org/wiki/Germany": {
			TotalLinkCount:    3708,
			ExternalLinkCount: 878,
			InternalLinkCount: 2830,
		},
		"https://web.archive.org/web/20220426164538/https://en.wikipedia.org/wiki/Go_(programming_language)": {
			TotalLinkCount:    1231,
			ExternalLinkCount: 254,
			InternalLinkCount: 977,
		},
	}

	for url, expectedReport := range expectedLinkCounts {
		report := InspectURL(url)
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
