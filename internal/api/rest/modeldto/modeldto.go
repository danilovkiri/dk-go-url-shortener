// Package modeldto provides locally used types and their structure for data transfer objects.
package modeldto

type (
	RequestURL struct {
		URL string `json:"url"`
	}

	ResponseURL struct {
		ShortURL string `json:"result"`
	}
)
