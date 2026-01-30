package contract

import "encoding/json"

// JSONRaw wraps json.RawMessage so swagger can resolve the type locally.
type JSONRaw json.RawMessage

func (j JSONRaw) MarshalJSON() ([]byte, error) {
	return json.RawMessage(j).MarshalJSON()
}

func (j *JSONRaw) UnmarshalJSON(b []byte) error {
	if j == nil {
		return nil
	}
	*j = JSONRaw(append([]byte(nil), b...))
	return nil
}

func RawMessagePtr(value *JSONRaw) *json.RawMessage {
	if value == nil {
		return nil
	}
	raw := json.RawMessage(*value)
	return &raw
}
