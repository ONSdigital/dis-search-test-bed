package settings

import _ "embed"

//go:embed search-index-settings.json
var searchIndexSettingsJSON []byte

// GetSearchIndexSettings returns the embedded JSON settings for the
// search index.
func GetSearchIndexSettings() []byte {
	return searchIndexSettingsJSON
}
