package algorithm

import (
	"bytes"
	"encoding/json"
)

// MinifyJSON returns rawJSON with insignificant whitespace removed.
func MinifyJSON(rawJSON []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, rawJSON); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
