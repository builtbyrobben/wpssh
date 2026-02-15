package cmd

// Globals holds flags that are available to every command.
type Globals struct {
	Site           string   `help:"Target site alias" short:"s" env:"WPGO_SITE"`
	Sites          []string `help:"Multiple target sites (batch mode)" sep:","`
	Group          string   `help:"Target site group" short:"g"`
	JSON           bool     `help:"Output as JSON" xor:"format"`
	Plain          bool     `help:"Output as plain text" xor:"format"`
	Verbose        bool     `help:"Verbose output" short:"v"`
	DryRun         bool     `help:"Show commands without executing" name:"dry-run"`
	NoCache        bool     `help:"Bypass cache" name:"no-cache"`
	Fields         string   `help:"Comma-separated fields to display"`
	Yes            bool     `help:"Skip confirmation prompts" short:"y"`
	AckDestructive bool     `help:"Acknowledge destructive batch operations" name:"ack-destructive"`
	Concurrency    int      `help:"Max parallel executions in batch mode" default:"1"`
}

// IsBatchMode returns true if the command targets multiple sites.
func (g *Globals) IsBatchMode() bool {
	return len(g.Sites) > 0 || g.Group != ""
}
