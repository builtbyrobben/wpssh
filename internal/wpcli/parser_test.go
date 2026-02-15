package wpcli

import (
	"testing"
)

func TestParseJSONPlugins(t *testing.T) {
	input := `[{"name":"akismet","status":"active","update":"none","version":"5.3","auto_update":"on"},{"name":"hello-dolly","status":"inactive","update":"available","version":"1.7.2","auto_update":"off"}]`

	plugins, err := ParseJSON[Plugin](input)
	if err != nil {
		t.Fatalf("ParseJSON error: %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("got %d plugins, want 2", len(plugins))
	}
	if plugins[0].Name != "akismet" {
		t.Errorf("plugins[0].Name = %q, want %q", plugins[0].Name, "akismet")
	}
	if plugins[0].Status != "active" {
		t.Errorf("plugins[0].Status = %q, want %q", plugins[0].Status, "active")
	}
	if plugins[1].AutoUpdate != "off" {
		t.Errorf("plugins[1].AutoUpdate = %q, want %q", plugins[1].AutoUpdate, "off")
	}
}

func TestParseJSONWithWarnings(t *testing.T) {
	input := `Warning: Some warning message
PHP Notice: something in /path/to/file.php on line 42
[{"name":"akismet","status":"active","update":"none","version":"5.3","auto_update":"on"}]`

	plugins, err := ParseJSON[Plugin](input)
	if err != nil {
		t.Fatalf("ParseJSON with warnings error: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("got %d plugins, want 1", len(plugins))
	}
	if plugins[0].Name != "akismet" {
		t.Errorf("Name = %q, want %q", plugins[0].Name, "akismet")
	}
}

func TestParseJSONEmpty(t *testing.T) {
	plugins, err := ParseJSON[Plugin]("")
	if err != nil {
		t.Fatalf("ParseJSON empty error: %v", err)
	}
	if plugins != nil {
		t.Errorf("expected nil for empty input, got %v", plugins)
	}
}

func TestParseJSONEmptyArray(t *testing.T) {
	plugins, err := ParseJSON[Plugin]("[]")
	if err != nil {
		t.Fatalf("ParseJSON empty array error: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestParseJSONUsers(t *testing.T) {
	input := `[{"ID":1,"user_login":"admin","display_name":"Admin","user_email":"admin@example.com","roles":"administrator","user_registered":"2024-01-01 00:00:00"}]`

	users, err := ParseJSON[User](input)
	if err != nil {
		t.Fatalf("ParseJSON error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("got %d users, want 1", len(users))
	}
	if users[0].ID != 1 {
		t.Errorf("ID = %d, want 1", users[0].ID)
	}
	if users[0].Roles != "administrator" {
		t.Errorf("Roles = %q, want %q", users[0].Roles, "administrator")
	}
}

func TestParseSingle(t *testing.T) {
	input := `{"name":"akismet","status":"active","update":"none","version":"5.3","auto_update":"on"}`

	plugin, err := ParseSingle[Plugin](input)
	if err != nil {
		t.Fatalf("ParseSingle error: %v", err)
	}
	if plugin.Name != "akismet" {
		t.Errorf("Name = %q, want %q", plugin.Name, "akismet")
	}
}

func TestParseSingleWithWarnings(t *testing.T) {
	input := `Warning: Some PHP notice
{"name":"akismet","status":"active","update":"none","version":"5.3","auto_update":"on"}`

	plugin, err := ParseSingle[Plugin](input)
	if err != nil {
		t.Fatalf("ParseSingle with warnings error: %v", err)
	}
	if plugin.Name != "akismet" {
		t.Errorf("Name = %q, want %q", plugin.Name, "akismet")
	}
}

func TestParseSingleEmpty(t *testing.T) {
	_, err := ParseSingle[Plugin]("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseJSONOnlyWarnings(t *testing.T) {
	input := "Warning: some message\nPHP Notice: another message"
	plugins, err := ParseJSON[Plugin](input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plugins != nil {
		t.Errorf("expected nil, got %v", plugins)
	}
}

func TestStripWarnings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no warnings",
			input: `[{"name":"test"}]`,
			want:  `[{"name":"test"}]`,
		},
		{
			name:  "single warning",
			input: "Warning: test\n[{\"name\":\"test\"}]",
			want:  `[{"name":"test"}]`,
		},
		{
			name:  "multiple warnings",
			input: "Warning: one\nNotice: two\n[{\"name\":\"test\"}]",
			want:  `[{"name":"test"}]`,
		},
		{
			name:  "object output",
			input: "Warning: test\n{\"name\":\"test\"}",
			want:  `{"name":"test"}`,
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "only warnings",
			input: "Warning: one\nNotice: two",
			want:  "",
		},
		{
			name:  "whitespace before json",
			input: "Warning: test\n  [{\"name\":\"test\"}]",
			want:  "  [{\"name\":\"test\"}]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripWarnings(tt.input)
			if got != tt.want {
				t.Errorf("stripWarnings() = %q, want %q", got, tt.want)
			}
		})
	}
}
