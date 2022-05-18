// Package modelstorage provides locally used types and their structure for storage objects.
package modelstorage

type URLStorageEntry struct {
	SURL   string `json:"sURL"`
	URL    string `json:"URL"`
	UserID string `json:"userID"`
}

type URLMapEntry struct {
	URL    string
	UserID string
}
