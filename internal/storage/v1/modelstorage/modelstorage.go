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

type URLPostgresEntry struct {
	ID        uint   `db:"id"`
	UserID    string `db:"user_id"` // store as a string since we store encoded tokens
	URL       string `db:"url"`
	SURL      string `db:"short_url"`
	IsDeleted bool   `db:"is_deleted"`
}

type URLChannelEntry struct {
	UserID string
	SURL   string
}
