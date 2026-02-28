package safety

import "testing"

func TestClassifyReadCommands(t *testing.T) {
	tests := []struct {
		parts []string
	}{
		{[]string{"plugin", "list"}},
		{[]string{"plugin", "get"}},
		{[]string{"plugin", "search"}},
		{[]string{"plugin", "status"}},
		{[]string{"plugin", "is-active"}},
		{[]string{"plugin", "is-installed"}},
		{[]string{"plugin", "verify-checksums"}},
		{[]string{"theme", "list"}},
		{[]string{"theme", "get"}},
		{[]string{"theme", "search"}},
		{[]string{"theme", "is-active"}},
		{[]string{"theme", "is-installed"}},
		{[]string{"core", "version"}},
		{[]string{"core", "check-update"}},
		{[]string{"core", "verify-checksums"}},
		{[]string{"core", "is-installed"}},
		{[]string{"user", "list"}},
		{[]string{"user", "get"}},
		{[]string{"user", "exists"}},
		{[]string{"post", "list"}},
		{[]string{"post", "get"}},
		{[]string{"post", "exists"}},
		{[]string{"option", "list"}},
		{[]string{"option", "get"}},
		{[]string{"option", "pluck"}},
		{[]string{"db", "size"}},
		{[]string{"db", "tables"}},
		{[]string{"db", "prefix"}},
		{[]string{"db", "check"}},
		{[]string{"db", "export"}},
		{[]string{"comment", "list"}},
		{[]string{"comment", "get"}},
		{[]string{"comment", "count"}},
		{[]string{"cron", "event", "list"}},
		{[]string{"cron", "schedule", "list"}},
		{[]string{"cron", "test"}},
		{[]string{"cache", "type"}},
		{[]string{"rewrite", "list"}},
		{[]string{"role", "list"}},
		{[]string{"role", "exists"}},
		{[]string{"maintenance", "status"}},
		{[]string{"config", "get"}},
		{[]string{"config", "has"}},
		{[]string{"config", "list"}},
		{[]string{"config", "path"}},
		{[]string{"media", "image-size"}},
		{[]string{"menu", "list"}},
		{[]string{"menu", "item", "list"}},
		// Top-level read shortcuts
		{[]string{"health"}},
		{[]string{"status"}},
		{[]string{"version"}},
	}

	for _, tt := range tests {
		tier := Classify(tt.parts...)
		if tier != TierRead {
			t.Errorf("Classify(%v) = %v, want TierRead", tt.parts, tier)
		}
	}
}

func TestClassifyMutateCommands(t *testing.T) {
	tests := []struct {
		parts []string
	}{
		{[]string{"plugin", "install"}},
		{[]string{"plugin", "activate"}},
		{[]string{"plugin", "deactivate"}},
		{[]string{"plugin", "update"}},
		{[]string{"plugin", "auto-updates"}},
		{[]string{"theme", "install"}},
		{[]string{"theme", "activate"}},
		{[]string{"theme", "update"}},
		{[]string{"core", "update"}},
		{[]string{"user", "create"}},
		{[]string{"user", "update"}},
		{[]string{"user", "set-role"}},
		{[]string{"user", "reset-password"}},
		{[]string{"post", "create"}},
		{[]string{"post", "update"}},
		{[]string{"option", "update"}},
		{[]string{"option", "add"}},
		{[]string{"option", "patch"}},
		{[]string{"cache", "flush"}},
		{[]string{"cache", "set"}},
		{[]string{"transient", "set"}},
		{[]string{"media", "regenerate"}},
		{[]string{"rewrite", "flush"}},
		{[]string{"rewrite", "structure"}},
		{[]string{"comment", "approve"}},
		{[]string{"comment", "unapprove"}},
		{[]string{"comment", "spam"}},
		{[]string{"comment", "unspam"}},
		{[]string{"comment", "trash"}},
		{[]string{"comment", "untrash"}},
		{[]string{"comment", "create"}},
		{[]string{"comment", "update"}},
		{[]string{"cron", "event", "run"}},
		{[]string{"cron", "event", "schedule"}},
		{[]string{"cron", "event", "unschedule"}},
		{[]string{"menu", "create"}},
		{[]string{"menu", "item", "add-post"}},
		{[]string{"menu", "item", "add-custom"}},
		{[]string{"role", "create"}},
		{[]string{"config", "set"}},
		{[]string{"maintenance", "activate"}},
		{[]string{"maintenance", "deactivate"}},
		// Top-level mutate shortcuts
		{[]string{"backup"}},
		{[]string{"clear-cache"}},
		{[]string{"update-all"}},
	}

	for _, tt := range tests {
		tier := Classify(tt.parts...)
		if tier != TierMutate {
			t.Errorf("Classify(%v) = %v, want TierMutate", tt.parts, tier)
		}
	}
}

func TestClassifyDestructiveCommands(t *testing.T) {
	tests := []struct {
		parts []string
	}{
		{[]string{"plugin", "delete"}},
		{[]string{"theme", "delete"}},
		{[]string{"user", "delete"}},
		{[]string{"post", "delete"}},
		{[]string{"option", "delete"}},
		{[]string{"comment", "delete"}},
		{[]string{"menu", "delete"}},
		{[]string{"menu", "item", "delete"}},
		{[]string{"role", "delete"}},
		{[]string{"config", "delete"}},
		{[]string{"cron", "event", "delete"}},
		{[]string{"transient", "delete"}},
		{[]string{"cache", "delete"}},
		{[]string{"db", "reset"}},
		{[]string{"db", "import"}},
		{[]string{"db", "repair"}},
		{[]string{"db", "optimize"}},
		// Top-level destructive commands
		{[]string{"eval"}},
		{[]string{"search-replace"}},
		{[]string{"raw"}},
	}

	for _, tt := range tests {
		tier := Classify(tt.parts...)
		if tier != TierDestructive {
			t.Errorf("Classify(%v) = %v, want TierDestructive", tt.parts, tier)
		}
	}
}

func TestClassifyEmpty(t *testing.T) {
	tier := Classify()
	if tier != TierRead {
		t.Errorf("Classify() = %v, want TierRead", tier)
	}
}

func TestSafetyTierString(t *testing.T) {
	tests := []struct {
		tier SafetyTier
		want string
	}{
		{TierRead, "read"},
		{TierMutate, "mutate"},
		{TierDestructive, "destructive"},
		{SafetyTier(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.tier.String()
		if got != tt.want {
			t.Errorf("SafetyTier(%d).String() = %q, want %q", tt.tier, got, tt.want)
		}
	}
}

func TestCheckBatchSafetyReadAlwaysAllowed(t *testing.T) {
	// Read tier passes with any combination of flags.
	for _, tc := range []struct {
		yes bool
		ack bool
	}{
		{false, false},
		{true, false},
		{false, true},
		{true, true},
	} {
		err := CheckBatchSafety(TierRead, tc.yes, tc.ack)
		if err != nil {
			t.Errorf("CheckBatchSafety(Read, yes=%v, ack=%v) = %v, want nil", tc.yes, tc.ack, err)
		}
	}
}

func TestCheckBatchSafetyMutateWithYes(t *testing.T) {
	err := CheckBatchSafety(TierMutate, true, false)
	if err != nil {
		t.Errorf("CheckBatchSafety(Mutate, yes=true) = %v, want nil", err)
	}
}

func TestCheckBatchSafetyMutateNonTTYWithoutYes(t *testing.T) {
	// In CI / non-TTY environments, this should fail.
	// Note: the test runner is typically not a TTY, so IsTTY() returns false.
	if IsTTY() {
		t.Skip("test requires non-TTY environment")
	}

	err := CheckBatchSafety(TierMutate, false, false)
	if err == nil {
		t.Error("CheckBatchSafety(Mutate, yes=false) in non-TTY should return error")
	}
	safetyErr, ok := err.(*ErrSafetyCheck)
	if !ok {
		t.Fatalf("error type = %T, want *ErrSafetyCheck", err)
	}
	if safetyErr.Tier != TierMutate {
		t.Errorf("error tier = %v, want TierMutate", safetyErr.Tier)
	}
}

func TestCheckBatchSafetyDestructiveRequiresBothFlags(t *testing.T) {
	tests := []struct {
		yes     bool
		ack     bool
		wantErr bool
	}{
		{false, false, true},
		{true, false, true},
		{false, true, true},
		{true, true, false},
	}

	for _, tc := range tests {
		err := CheckBatchSafety(TierDestructive, tc.yes, tc.ack)
		if tc.wantErr && err == nil {
			t.Errorf("CheckBatchSafety(Destructive, yes=%v, ack=%v) = nil, want error", tc.yes, tc.ack)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("CheckBatchSafety(Destructive, yes=%v, ack=%v) = %v, want nil", tc.yes, tc.ack, err)
		}
	}
}

func TestCheckBatchSafetyDestructiveErrorType(t *testing.T) {
	err := CheckBatchSafety(TierDestructive, false, false)
	if err == nil {
		t.Fatal("expected error")
	}
	safetyErr, ok := err.(*ErrSafetyCheck)
	if !ok {
		t.Fatalf("error type = %T, want *ErrSafetyCheck", err)
	}
	if safetyErr.Tier != TierDestructive {
		t.Errorf("error tier = %v, want TierDestructive", safetyErr.Tier)
	}
}

func TestCheckBatchSafetyDestructiveYesWithoutAck(t *testing.T) {
	err := CheckBatchSafety(TierDestructive, true, false)
	if err == nil {
		t.Error("expected error: --yes without --ack-destructive should fail for destructive")
	}
	safetyErr, ok := err.(*ErrSafetyCheck)
	if !ok {
		t.Fatalf("error type = %T, want *ErrSafetyCheck", err)
	}
	if safetyErr.Tier != TierDestructive {
		t.Errorf("error tier = %v, want TierDestructive", safetyErr.Tier)
	}
}
