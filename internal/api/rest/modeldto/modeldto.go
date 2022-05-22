// Package modeldto provides locally used types and their structure for data transfer objects.
package modeldto

type (
	// RequestURL is used in JSONHandlePostURL
	RequestURL struct {
		URL string `json:"url"`
	}

	// ResponseURL is used in JSONHandlePostURL
	ResponseURL struct {
		SURL string `json:"result"`
	}

	// ResponseFullURL is used in HandleGetURLsByUserID
	ResponseFullURL struct {
		URL  string `json:"original_url"`
		SURL string `json:"short_url"`
	}

	// RequestBatchURL is used in JSONHandlePostURLBatch
	RequestBatchURL struct {
		CorrelationID string `json:"correlation_id"`
		URL           string `json:"original_url"`
	}

	// ResponseBatchURL is used in JSONHandlePostURLBatch
	ResponseBatchURL struct {
		CorrelationID string `json:"correlation_id"`
		SURL          string `json:"short_url"`
	}
)
