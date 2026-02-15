package outfmt

import (
	"testing"
)

type testPlugin struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

func TestProjectFieldsEmpty(t *testing.T) {
	data := testPlugin{Name: "akismet", Status: "active", Version: "5.3"}
	result, err := projectFields(data, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return data unchanged
	p, ok := result.(testPlugin)
	if !ok {
		t.Fatal("expected testPlugin type")
	}
	if p.Name != "akismet" {
		t.Errorf("Name = %q, want %q", p.Name, "akismet")
	}
}

func TestProjectFieldsStruct(t *testing.T) {
	data := testPlugin{Name: "akismet", Status: "active", Version: "5.3"}
	result, err := projectFields(data, "name,version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}
	if m["name"] != "akismet" {
		t.Errorf("name = %v, want %q", m["name"], "akismet")
	}
	if m["version"] != "5.3" {
		t.Errorf("version = %v, want %q", m["version"], "5.3")
	}
	if _, exists := m["status"]; exists {
		t.Error("status should not be in projected result")
	}
}

func TestProjectFieldsSlice(t *testing.T) {
	data := []testPlugin{
		{Name: "akismet", Status: "active", Version: "5.3"},
		{Name: "hello-dolly", Status: "inactive", Version: "1.7.2"},
	}
	result, err := projectFields(data, "name,status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	slice, ok := result.([]map[string]any)
	if !ok {
		t.Fatalf("expected []map[string]any, got %T", result)
	}
	if len(slice) != 2 {
		t.Fatalf("len = %d, want 2", len(slice))
	}
	if slice[0]["name"] != "akismet" {
		t.Errorf("[0].name = %v, want %q", slice[0]["name"], "akismet")
	}
	if slice[1]["status"] != "inactive" {
		t.Errorf("[1].status = %v, want %q", slice[1]["status"], "inactive")
	}
	if _, exists := slice[0]["version"]; exists {
		t.Error("version should not be in projected result")
	}
}

func TestProjectFieldsCaseInsensitive(t *testing.T) {
	data := testPlugin{Name: "akismet", Status: "active", Version: "5.3"}
	result, err := projectFields(data, "NAME,Version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}
	if m["name"] != "akismet" {
		t.Errorf("name = %v, want %q", m["name"], "akismet")
	}
	if m["version"] != "5.3" {
		t.Errorf("version = %v, want %q", m["version"], "5.3")
	}
}

func TestProjectFieldsMap(t *testing.T) {
	data := map[string]any{
		"name":    "akismet",
		"status":  "active",
		"version": "5.3",
	}
	result, err := projectFields(data, "name,version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}
	if m["name"] != "akismet" {
		t.Errorf("name = %v, want %q", m["name"], "akismet")
	}
	if _, exists := m["status"]; exists {
		t.Error("status should not be in projected result")
	}
}

func TestParseFieldList(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"name,version", 2},
		{"name, version, status", 3},
		{"", 0},
		{" name ", 1},
		{",,,", 0},
	}

	for _, tt := range tests {
		got := parseFieldList(tt.input)
		if len(got) != tt.want {
			t.Errorf("parseFieldList(%q) len = %d, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestProjectFieldsNil(t *testing.T) {
	result, err := projectFields(nil, "name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestProjectFieldsSliceOfMaps(t *testing.T) {
	data := []map[string]any{
		{"name": "akismet", "status": "active", "version": "5.3"},
		{"name": "hello-dolly", "status": "inactive", "version": "1.7.2"},
	}
	result, err := projectFields(data, "name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	slice, ok := result.([]map[string]any)
	if !ok {
		t.Fatalf("expected []map[string]any, got %T", result)
	}
	if len(slice) != 2 {
		t.Fatalf("len = %d, want 2", len(slice))
	}
	if slice[0]["name"] != "akismet" {
		t.Errorf("[0].name = %v, want %q", slice[0]["name"], "akismet")
	}
	if _, exists := slice[0]["status"]; exists {
		t.Error("status should not be in projected result")
	}
}
