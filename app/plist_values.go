package app

import (
	"fmt"
	"sort"
)

type plistDict = map[string]interface{}

func plistString(m plistDict, key string) string {
	switch s := m[key].(type) {
	case string:
		return s
	case uint64, int64, float64, bool:
		return fmt.Sprint(s)
	}
	return ""
}

// plistStrings reads a key that holds a list of strings, a dict of
// capability: required (UIRequiredDeviceCapabilities) or a single string
func plistStrings(m plistDict, key string) []string {
	switch v := m[key].(type) {
	case []interface{}:
		return stringsOfList(v)
	case plistDict:
		return requiredKeys(v)
	case string:
		return []string{v}
	}
	return nil
}

func stringsOfList(list []interface{}) []string {
	var out []string
	for _, e := range list {
		switch s := e.(type) {
		case string:
			out = append(out, s)
		case uint64, int64:
			out = append(out, fmt.Sprint(s))
		}
	}
	return out
}

// requiredKeys are the keys of a dict whose value is not false, sorted
func requiredKeys(m plistDict) []string {
	var out []string
	for k, b := range m {
		if req, ok := b.(bool); !ok || req {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
