package testhelpers

import "regexp"

func GetEmailFromBody(reqBody string) string {
	re := regexp.MustCompile(`"email"\s*:\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(reqBody)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func GetPasswordFromBody(reqBody string) string {
	re := regexp.MustCompile(`"password"\s*:\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(reqBody)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
