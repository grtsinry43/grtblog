package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
)

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

func jsonRawFromBytes(value []byte) *contract.JSONRaw {
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}
	copied := append([]byte(nil), trimmed...)
	raw := contract.JSONRaw(copied)
	return &raw
}
