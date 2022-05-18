// Package modeldto provides locally used types and their structure for data transfer objects.
package modeldto

type (
	RequestURL struct {
		URL string `json:"url"`
	}

	ResponseURL struct {
		SURL string `json:"result"`
	}

	ResponseFullURL struct {
		URL  string `json:"original_url"`
		SURL string `json:"short_url"`
	}
)
