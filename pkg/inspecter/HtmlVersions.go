package inspecter

import "strings"

var HTMLVersions = map[string]string{
	`XHTML 1.1`:              `/dtd xhtml 1.1/`,
	`XHTML 1.0 Strict`:       `/dtd xhtml 1.0 strict/`,
	`XHTML 1.0 Transitional`: `/dtd xhtml 1.0 transitional/`,
	`XHTML 1.0 Frameset`:     `/dtd xhtml 1.0 frameset/`,
	`HTML 4.01 Strict`:       `/dtd html 4.01/`,
	`HTML 4.01 Transitional`: `/dtd html 4.01 transitional/`,
	`HTML 4.01 Frameset`:     `/dtd html 4.01 frameset/`,
	`HTML 5`:                 `html`,
}

// DetectHTMLVersion returns the HTML version of the given HTML Doctype Tag
func DetectHTMLVersion(doctypeTag string) string {
	if len(doctypeTag) == 0 {
		return `Not defined`
	}

	// Normalize the doctype tag
	doctypeTag = strings.ToLower(doctypeTag)
	doctypeTag = strings.ReplaceAll(doctypeTag, `  `, ` `)

	// Check if the doctype tag is in the map
	for k, v := range HTMLVersions {
		if strings.Contains(doctypeTag, v) {
			return k
		}
	}

	return doctypeTag
}
