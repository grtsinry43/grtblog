package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
)

var momentWeatherValues = map[string]struct{}{
	"sunny":    {},
	"cloudy":   {},
	"overcast": {},
	"rainy":    {},
	"snowy":    {},
	"windy":    {},
	"foggy":    {},
}

var momentMoodValues = map[string]struct{}{
	"joyful":  {},
	"calm":    {},
	"excited": {},
	"tired":   {},
	"sad":     {},
}

func parseExtInfo(raw *contract.JSONRaw) ([]byte, error) {
	if raw == nil {
		return nil, nil
	}
	data := []byte(*raw)
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		return nil, err
	}
	if images, ok := payload["images"]; ok && images != nil {
		items, ok := images.([]any)
		if !ok {
			return nil, errors.New("images must be array")
		}
		for _, item := range items {
			obj, ok := item.(map[string]any)
			if !ok {
				return nil, errors.New("images item must be object")
			}
			idRaw, ok := obj["id"]
			if !ok {
				return nil, errors.New("images item missing id")
			}
			id, ok := idRaw.(string)
			if !ok || strings.TrimSpace(id) == "" {
				return nil, errors.New("images item id must be non-empty string")
			}
		}
	}

	return append([]byte(nil), trimmed...), nil
}

func parseMomentExtInfo(raw *contract.JSONRaw) ([]byte, error) {
	data, err := parseExtInfo(raw)
	if err != nil || len(data) == 0 {
		return data, err
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	momentRaw, ok := payload["moment"]
	if !ok || bytes.Equal(bytes.TrimSpace(momentRaw), []byte("null")) {
		return data, nil
	}

	var momentInfo map[string]json.RawMessage
	if err := json.Unmarshal(momentRaw, &momentInfo); err != nil {
		return nil, errors.New("moment must be object")
	}
	if momentInfo == nil {
		return nil, errors.New("moment must be object")
	}
	if err := validateMomentExtInfoValue(momentInfo, "weather", momentWeatherValues); err != nil {
		return nil, err
	}
	if err := validateMomentExtInfoValue(momentInfo, "mood", momentMoodValues); err != nil {
		return nil, err
	}

	return data, nil
}

func validateMomentExtInfoValue(
	momentInfo map[string]json.RawMessage,
	field string,
	allowed map[string]struct{},
) error {
	raw, ok := momentInfo[field]
	if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("moment.%s must be string", field)
	}
	if _, ok := allowed[value]; !ok {
		return fmt.Errorf("moment.%s has unsupported value %q", field, value)
	}
	return nil
}

func jsonRawFromBytes(value []byte) *contract.JSONRaw {
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}
	copied := append([]byte(nil), trimmed...)
	raw := contract.JSONRaw(copied)
	return &raw
}
