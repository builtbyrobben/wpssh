package outfmt

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// formatPlain writes data as tab-separated values (no headers) for piping.
func formatPlain(w io.Writer, data any) error {
	rows, _ := toRows(data)
	for _, row := range rows {
		vals := make([]string, 0, len(row))
		for _, v := range row {
			vals = append(vals, fmt.Sprint(v))
		}
		fmt.Fprintln(w, strings.Join(vals, "\t"))
	}
	return nil
}

// toRows converts data to a list of ordered value slices.
// Returns (rows, headers).
func toRows(data any) ([][]any, []string) {
	v := reflect.ValueOf(data)

	switch v.Kind() {
	case reflect.Slice:
		if v.Len() == 0 {
			return nil, nil
		}
		first := v.Index(0)
		if first.Kind() == reflect.Interface {
			first = first.Elem()
		}
		headers := getHeaders(first)
		rows := make([][]any, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Interface {
				elem = elem.Elem()
			}
			rows = append(rows, getValues(elem, headers))
		}
		return rows, headers

	case reflect.Struct:
		headers := getHeaders(v)
		return [][]any{getValues(v, headers)}, headers

	case reflect.Map:
		headers := getMapHeaders(v)
		return [][]any{getMapValues(v, headers)}, headers

	case reflect.Ptr:
		if v.IsNil() {
			return nil, nil
		}
		return toRows(v.Elem().Interface())

	default:
		return [][]any{{data}}, nil
	}
}

// getHeaders returns column names from a struct or map.
func getHeaders(v reflect.Value) []string {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return getStructHeaders(v)
	case reflect.Map:
		return getMapHeaders(v)
	default:
		return nil
	}
}

// getStructHeaders returns JSON tag names (or field names) for exported fields.
func getStructHeaders(v reflect.Value) []string {
	t := v.Type()
	headers := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name := f.Name
		if tag != "" {
			name = strings.Split(tag, ",")[0]
		}
		headers = append(headers, name)
	}
	return headers
}

// getMapHeaders returns sorted keys from a map.
func getMapHeaders(v reflect.Value) []string {
	keys := v.MapKeys()
	headers := make([]string, 0, len(keys))
	for _, k := range keys {
		headers = append(headers, fmt.Sprint(k.Interface()))
	}
	// Sort for deterministic output
	sortStrings(headers)
	return headers
}

// getValues extracts values from a struct in header order.
func getValues(v reflect.Value, headers []string) []any {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return getStructValues(v, headers)
	case reflect.Map:
		return getMapValues(v, headers)
	default:
		return []any{v.Interface()}
	}
}

// getStructValues returns field values matching headers by JSON tag or name.
func getStructValues(v reflect.Value, headers []string) []any {
	t := v.Type()

	// Build lookup: header name -> field index
	lookup := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name := f.Name
		if tag != "" {
			name = strings.Split(tag, ",")[0]
		}
		lookup[name] = i
	}

	vals := make([]any, 0, len(headers))
	for _, h := range headers {
		if idx, ok := lookup[h]; ok {
			vals = append(vals, v.Field(idx).Interface())
		} else {
			vals = append(vals, "")
		}
	}
	return vals
}

// getMapValues returns map values in header order.
func getMapValues(v reflect.Value, headers []string) []any {
	vals := make([]any, 0, len(headers))
	for _, h := range headers {
		val := v.MapIndex(reflect.ValueOf(h))
		if val.IsValid() {
			vals = append(vals, val.Interface())
		} else {
			vals = append(vals, "")
		}
	}
	return vals
}

// sortStrings sorts a string slice in place.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
