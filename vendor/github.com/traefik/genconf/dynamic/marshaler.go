package dynamic

import "encoding/json"

// JSONPayload holds the dynamic configuration to marshal as JSON.
type JSONPayload struct {
	*Configuration
}

// MarshalJSON implements json.Marshaler.
func (c JSONPayload) MarshalJSON() ([]byte, error) {
	if c.Configuration == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(c.Configuration)
}
