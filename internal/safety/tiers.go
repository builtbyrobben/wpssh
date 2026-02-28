package safety

import (
	"fmt"
	"os"
)

// SafetyTier classifies wp-cli commands by their potential impact.
type SafetyTier int

const (
	// TierRead commands have no side effects and need no confirmation.
	TierRead SafetyTier = iota

	// TierMutate commands modify state. In batch mode, they require --yes
	// or an interactive TTY prompt. Non-TTY without --yes exits with code 1.
	TierMutate

	// TierDestructive commands can cause irreversible data loss. In batch mode,
	// they require both --yes AND --ack-destructive. No interactive fallback.
	TierDestructive
)

// String returns a human-readable name for the safety tier.
func (t SafetyTier) String() string {
	switch t {
	case TierRead:
		return "read"
	case TierMutate:
		return "mutate"
	case TierDestructive:
		return "destructive"
	default:
		return "unknown"
	}
}

// readSubcommands are subcommands classified as read-only regardless of parent command.
var readSubcommands = map[string]bool{
	"list":             true,
	"get":              true,
	"search":           true,
	"status":           true,
	"version":          true,
	"check-update":     true,
	"verify-checksums": true,
	"is-active":        true,
	"is-installed":     true,
	"exists":           true,
	"size":             true,
	"tables":           true,
	"prefix":           true,
	"count":            true,
	"type":             true,
	"path":             true,
	"health":           true,
	"check":            true,
	"export":           true,
	"has":              true,
	"pluck":            true,
	"test":             true,
	"image-size":       true,
}

// mutateSubcommands are subcommands classified as state-mutating.
var mutateSubcommands = map[string]bool{
	"install":        true,
	"activate":       true,
	"deactivate":     true,
	"update":         true,
	"create":         true,
	"set":            true,
	"add":            true,
	"flush":          true,
	"regenerate":     true,
	"approve":        true,
	"unapprove":      true,
	"spam":           true,
	"unspam":         true,
	"trash":          true,
	"untrash":        true,
	"schedule":       true,
	"unschedule":     true,
	"structure":      true,
	"auto-updates":   true,
	"set-role":       true,
	"reset-password": true,
	"patch":          true,
	"add-post":       true,
	"add-custom":     true,
	"run":            true,
}

// destructiveSubcommands are subcommands classified as potentially irreversible.
var destructiveSubcommands = map[string]bool{
	"delete":         true,
	"reset":          true,
	"search-replace": true,
	"repair":         true,
	"optimize":       true,
}

// destructiveOverrides maps specific "parent subcommand" combos to destructive tier,
// overriding the general subcommand classification.
var destructiveOverrides = map[string]bool{
	"db import": true,
}

// Classify returns the safety tier for a wp-cli command.
// Parts should be the command components, e.g., ("plugin", "list") or ("db", "import").
func Classify(parts ...string) SafetyTier {
	if len(parts) == 0 {
		return TierRead
	}

	// Top-level commands that are always destructive
	topLevel := parts[0]
	if topLevel == "eval" || topLevel == "search-replace" {
		return TierDestructive
	}

	// Top-level shortcuts
	switch topLevel {
	case "health", "status", "version":
		return TierRead
	case "backup", "clear-cache", "update-all":
		return TierMutate
	case "raw":
		return TierDestructive
	}

	if len(parts) < 2 {
		return TierRead
	}

	subcommand := parts[len(parts)-1]

	// Check specific overrides first (e.g., "db import")
	combined := topLevel + " " + subcommand
	if destructiveOverrides[combined] {
		return TierDestructive
	}

	// Check subcommand classification
	if destructiveSubcommands[subcommand] {
		return TierDestructive
	}

	if mutateSubcommands[subcommand] {
		return TierMutate
	}

	if readSubcommands[subcommand] {
		return TierRead
	}

	// Default to mutate for unknown subcommands (safer than read)
	return TierMutate
}

// IsTTY returns true if stdin is connected to a terminal.
func IsTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// ErrSafetyCheck is returned when a batch safety gate blocks execution.
type ErrSafetyCheck struct {
	Tier    SafetyTier
	Message string
}

func (e *ErrSafetyCheck) Error() string {
	return e.Message
}

// CheckBatchSafety validates that the user has provided the required
// flags for the given safety tier in batch mode. Returns nil for read
// commands. Returns an ErrSafetyCheck if the operation is blocked.
//
// Rules:
//   - TierRead: always allowed
//   - TierMutate: requires --yes flag or interactive TTY. Non-TTY without
//     --yes returns an error (exit code 1).
//   - TierDestructive: requires BOTH --yes AND --ack-destructive.
//     No interactive fallback.
func CheckBatchSafety(tier SafetyTier, yes, ackDestructive bool) error {
	switch tier {
	case TierRead:
		return nil

	case TierMutate:
		if yes {
			return nil
		}
		if IsTTY() {
			// Caller should prompt interactively -- this function signals
			// that prompting is needed by returning nil (the caller is
			// responsible for the interactive prompt).
			return nil
		}
		return &ErrSafetyCheck{
			Tier:    TierMutate,
			Message: fmt.Sprintf("batch %s operation requires --yes flag in non-interactive mode", tier),
		}

	case TierDestructive:
		if !yes {
			return &ErrSafetyCheck{
				Tier:    TierDestructive,
				Message: fmt.Sprintf("batch %s operation requires --yes flag", tier),
			}
		}
		if !ackDestructive {
			return &ErrSafetyCheck{
				Tier:    TierDestructive,
				Message: fmt.Sprintf("batch %s operation requires --ack-destructive flag", tier),
			}
		}
		return nil

	default:
		return nil
	}
}

// NeedsPrompt returns true if the tier+flags combination requires
// an interactive confirmation prompt (only for TierMutate in TTY
// mode without --yes).
func NeedsPrompt(tier SafetyTier, yes bool) bool {
	return tier == TierMutate && !yes && IsTTY()
}
