package lambdahandler

import "regexp"

// buildPathPattern converts a path like "/resource/:id" into a regex pattern
func buildPathPattern(path string) *regexp.Regexp {
	pattern := regexp.MustCompile(`:([^/]+)`) // Match `:parameter`
	regexStr := "^" + pattern.ReplaceAllString(path, `([^/]+)`) + "$"
	return regexp.MustCompile(regexStr)
}

// extractParams extracts parameters from a matched path based on the regex
func extractParams(path string, pattern *regexp.Regexp) Params {
	matches := pattern.FindStringSubmatch(path)
	params := Params{}

	for i, name := range pattern.SubexpNames() {
		if i > 0 && name != "" { // Skip full match (index 0)
			params[name] = matches[i]
		}
	}

	return params
}
