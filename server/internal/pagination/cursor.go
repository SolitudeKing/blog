package pagination

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type CursorPayload struct {
	CreatedAt  time.Time `json:"created_at"`
	ID         uint64    `json:"id"`
	FilterHash string    `json:"filter_hash,omitempty"`
}

func EncodeCursor(payload CursorPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func DecodeCursor(cursor string) (CursorPayload, error) {
	var payload CursorPayload
	if cursor == "" {
		return payload, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return payload, err
	}
	err = json.Unmarshal(raw, &payload)
	return payload, err
}
