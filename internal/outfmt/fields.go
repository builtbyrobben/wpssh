package outfmt

import (
	"fmt"
	"reflect"
	"strings"
)

// projectFields filters data to only include the specified fields.
// If fields is empty, data is returned unchanged.
// Supports structs, slices of structs, and []map[string]any.
func projectFields(data any, fields string) (any, error) {
	if fields == "" {
		return data, nil
	}

	wanted := parseFieldList(fields)
	if len(wanted) == 0 {
		return data, nil
	}

	v := reflect.ValueOf(data)

	switch v.Kind() {
	case reflect.Slice:
		return projectSlice(v, wanted)
	case reflect.Struct:
		return projectStruct(v, wanted)
	case reflect.Map:
		return extractFromMap(v, wanted), nil
	case reflect.Ptr:
		if v.IsNil() {
			return data, nil
		}
		return projectFields(v.Elem().Interface(), fields)
	default:
		return data, nil
	}
}

// parseFieldList splits a comma-separated field string into lowercase field names.
func parseFieldList(fields string) []string {
	parts := strings.Split(fields, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, strings.ToLower(trimmed))
		}
	}
	return result
}

// projectSlice projects fields from each element in a slice.
func projectSlice(v reflect.Value, wanted []string) (any, error) {
	result := make([]map[string]any, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}
		m, err := extractFields(elem, wanted)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}

// projectStruct projects fields from a single struct.
func projectStruct(v reflect.Value, wanted []string) (map[string]any, error) {
	return extractFields(v, wanted)
}

// extractFields extracts the specified fields from a struct or map value.
func extractFields(v reflect.Value, wanted []string) (map[string]any, error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return extractFromStruct(v, wanted), nil
	case reflect.Map:
		return extractFromMap(v, wanted), nil
	default:
		return nil, fmt.Errorf("cannot extract fields from %s", v.Kind())
	}
}

// extractFromStruct extracts fields from a struct by matching JSON tags
// or field names (case-insensitive).
func extractFromStruct(v reflect.Value, wanted []string) map[string]any {
	t := v.Type()
	result := make(map[string]any, len(wanted))

	// Build field lookup: json tag name (lowercase) -> field index
	fieldMap := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		// Use JSON tag name if available
		tag := f.Tag.Get("json")
		if tag != "" && tag != "-" {
			name := strings.Split(tag, ",")[0]
			fieldMap[strings.ToLower(name)] = i
		}
		// Also index by field name
		fieldMap[strings.ToLower(f.Name)] = i
	}

	for _, w := range wanted {
		if idx, ok := fieldMap[w]; ok {
			f := t.Field(idx)
			tag := f.Tag.Get("json")
			key := f.Name
			if tag != "" && tag != "-" {
				key = strings.Split(tag, ",")[0]
			}
			result[key] = v.Field(idx).Interface()
		}
	}

	return result
}

// extractFromMap extracts fields from a map (case-insensitive key matching).
func extractFromMap(v reflect.Value, wanted []string) map[string]any {
	result := make(map[string]any, len(wanted))

	// Build lowercase key map
	keyMap := make(map[string]reflect.Value)
	for _, key := range v.MapKeys() {
		keyMap[strings.ToLower(fmt.Sprint(key.Interface()))] = key
	}

	for _, w := range wanted {
		if key, ok := keyMap[w]; ok {
			result[fmt.Sprint(key.Interface())] = v.MapIndex(key).Interface()
		}
	}

	return result
}
