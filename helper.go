package keyfile

import (
	"cmp"
	"reflect"
	"regexp"
	"slices"
	"strings"
)

func isOmitempty(tag reflect.StructTag) bool {
	tagField, ok := tag.Lookup(structTag)
	if !ok {
		return false
	}
	parts := split(tagField, ",")
	return slices.ContainsFunc(parts, func(part string) bool {
		return strings.TrimSpace(part) == "omitempty"
	})
}
func isIgnored(tag reflect.StructTag) bool {
	tagField, ok := tag.Lookup(structTag)
	if !ok {
		return false
	}
	parts := split(tagField, ",")
	return slices.ContainsFunc(parts, func(part string) bool {
		return strings.TrimSpace(part) == "-"
	})
}

func getKeyName(tag reflect.StructTag) string {
	tagField, ok := tag.Lookup(structTag)
	if !ok {
		return ""
	}
	parts := split(tagField, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if !strings.Contains(part, "=") && part != "-" && part != "omitempty" {
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
	sep := ""
	parts := split(tagField, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "sep=") {
			sep = strings.ReplaceAll(strings.TrimPrefix(part, "sep="), "\\", "")
			break
		}
	}

	return cmp.Or(sep, ";")
}

func split(value string, sep string) []string {
	result := make([]string, 0)
	buff := make([]rune, 0)
	for i, v := range value {
		if i == 0 && string(v) == sep {
			result = append(result, string(buff))
			buff = make([]rune, 0)
			continue
		}
		if string(v) == sep && value[i-1] != '\\' {
			result = append(result, string(buff))
			buff = make([]rune, 0)
			continue
		}
		buff = append(buff, v)
	}
	if len(buff) > 0 {
		result = append(result, string(buff))
	}
	return result
}

func unescape(value string) string {
	value = strings.ReplaceAll(value, "\\s", " ")
	value = strings.ReplaceAll(value, "\\n", "\n")
	value = strings.ReplaceAll(value, "\\r", "\r")
	value = strings.ReplaceAll(value, "\\t", "\t")
	return value
}

func replaceSpaces(input string) string {
	re := regexp.MustCompile(`^\s+|\s+$`)
	return re.ReplaceAllStringFunc(input, func(match string) string {
		return strings.Repeat("\\s", len(match))
	})
}

func escape(value string) string {
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\t", "\\t")
	value = replaceSpaces(value)
	return value
}
