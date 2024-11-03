package keyfile

import (
	"reflect"
	"slices"
	"strings"
)

func isIgnored(tag reflect.StructTag) bool {
	tagField, ok := tag.Lookup(structTag)
	if !ok {
		return false
	}
	parts := strings.Split(tagField, ";")
	return slices.ContainsFunc(parts, func(part string) bool {
		return strings.TrimSpace(part) == "-"
	})
}

func getKeyName(tag reflect.StructTag) string {
	tagField, ok := tag.Lookup(structTag)
	if !ok {
		return ""
	}
	parts := strings.Split(tagField, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if !strings.Contains(part, ":") && part != "-" && part != "omitempty" {
			return part
		}
	}
	return ""
}

func getSeperator(tag reflect.StructTag) string {
	tagField, ok := tag.Lookup(structTag)
	if !ok {
		return ""
	}
	parts := strings.Split(tagField, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if !strings.Contains(part, ":") {
			continue
		}
		kv := strings.Split(part, ":")
		if len(kv) != 2 {
			continue
		}
		if kv[0] == "sep" {
			return kv[1]
		}
	}
	return ""
}
