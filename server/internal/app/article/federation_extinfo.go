package article

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

const federationRegistryKey = "_federation_delivery_registry"

type federationRegistry struct {
	Mentions  map[string]string `json:"mentions,omitempty"`
	Citations map[string]string `json:"citations,omitempty"`
}

func mergeExtInfoKeepingFederation(base []byte, incoming []byte) []byte {
	baseObj := parseExtInfoObject(base)
	baseReg := registryFromObj(baseObj)

	// If caller didn't pass ext_info, preserve existing object as-is.
	if len(bytes.TrimSpace(incoming)) == 0 {
		return normalizeExtInfo(baseObj)
	}

	incomingObj := parseExtInfoObject(incoming)
	// Preserve existing image metadata when client sends partial ext_info.
	if _, ok := incomingObj["images"]; !ok {
		if images, exists := baseObj["images"]; exists {
			incomingObj["images"] = images
		}
	}
	incomingReg := registryFromObj(incomingObj)
	mergedReg := mergeRegistry(baseReg, incomingReg)
	setRegistry(incomingObj, mergedReg)
	return normalizeExtInfo(incomingObj)
}

func markDeliveredSignals(extInfo []byte, mentions []string, citations []string) ([]byte, bool) {
	obj := parseExtInfoObject(extInfo)
	reg := registryFromObj(obj)
	changed := false
	now := time.Now().UTC().Format(time.RFC3339)

	for _, key := range mentions {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		if _, ok := reg.Mentions[normalized]; ok {
			continue
		}
		reg.Mentions[normalized] = now
		changed = true
	}
	for _, key := range citations {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		if _, ok := reg.Citations[normalized]; ok {
			continue
		}
		reg.Citations[normalized] = now
		changed = true
	}
	if !changed {
		return extInfo, false
	}

	setRegistry(obj, reg)
	return normalizeExtInfo(obj), true
}

func deliveredSignalKeys(extInfo []byte) (map[string]struct{}, map[string]struct{}) {
	reg := registryFromObj(parseExtInfoObject(extInfo))
	mentionSet := make(map[string]struct{}, len(reg.Mentions))
	citationSet := make(map[string]struct{}, len(reg.Citations))
	for key := range reg.Mentions {
		mentionSet[key] = struct{}{}
	}
	for key := range reg.Citations {
		citationSet[key] = struct{}{}
	}
	return mentionSet, citationSet
}

func parseExtInfoObject(raw []byte) map[string]any {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return map[string]any{}
	}
	var obj map[string]any
	if err := json.Unmarshal(trimmed, &obj); err != nil {
		return map[string]any{}
	}
	if obj == nil {
		return map[string]any{}
	}
	return obj
}

func registryFromObj(obj map[string]any) federationRegistry {
	reg := federationRegistry{
		Mentions:  map[string]string{},
		Citations: map[string]string{},
	}
	raw, ok := obj[federationRegistryKey].(map[string]any)
	if !ok {
		return reg
	}
	copyStringMap(raw["mentions"], reg.Mentions)
	copyStringMap(raw["citations"], reg.Citations)
	return reg
}

func copyStringMap(src any, dst map[string]string) {
	raw, ok := src.(map[string]any)
	if !ok {
		return
	}
	for key, value := range raw {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		if str, ok := value.(string); ok && strings.TrimSpace(str) != "" {
			dst[normalized] = strings.TrimSpace(str)
			continue
		}
		dst[normalized] = ""
	}
}

func mergeRegistry(a federationRegistry, b federationRegistry) federationRegistry {
	merged := federationRegistry{
		Mentions:  map[string]string{},
		Citations: map[string]string{},
	}
	for key, ts := range a.Mentions {
		merged.Mentions[key] = ts
	}
	for key, ts := range b.Mentions {
		if _, ok := merged.Mentions[key]; !ok || ts != "" {
			merged.Mentions[key] = ts
		}
	}
	for key, ts := range a.Citations {
		merged.Citations[key] = ts
	}
	for key, ts := range b.Citations {
		if _, ok := merged.Citations[key]; !ok || ts != "" {
			merged.Citations[key] = ts
		}
	}
	return merged
}

func setRegistry(obj map[string]any, reg federationRegistry) {
	payload := map[string]any{
		"mentions":  reg.Mentions,
		"citations": reg.Citations,
	}
	obj[federationRegistryKey] = payload
}

func normalizeExtInfo(obj map[string]any) []byte {
	if len(obj) == 0 {
		return nil
	}
	// Marshal deterministic key order to avoid noisy updates.
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	ordered := make(map[string]any, len(obj))
	for _, key := range keys {
		ordered[key] = obj[key]
	}
	data, err := json.Marshal(ordered)
	if err != nil {
		return nil
	}
	return data
}
