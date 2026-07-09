package algorithm

import (
	"bytes"
	"encoding/json"
)

func minifyJSON(rawJSON []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, rawJSON); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
