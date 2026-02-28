package cmd

import (
	"fmt"

	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Transient commands

type TransientCmd struct {
	Get    TransientGetCmd    `cmd:"" help:"Get transient"`
	Set    TransientSetCmd    `cmd:"" help:"Set transient"`
	Delete TransientDeleteCmd `cmd:"" help:"Delete transient(s)"`
	List   TransientListCmd   `cmd:"" help:"List transients"`
}

type TransientGetCmd struct {
	Key string `arg:"" help:"Transient key"`
}
type TransientSetCmd struct {
	Key        string `arg:"" help:"Transient key"`
	Value      string `arg:"" help:"Transient value"`
	Expiration int    `arg:"" optional:"" help:"Expiration in seconds"`
}
type TransientDeleteCmd struct {
	Key     string `arg:"" optional:"" help:"Transient key"`
	All     bool   `help:"Delete all transients"`
	Expired bool   `help:"Delete expired transients"`
}
type TransientListCmd struct{}

func (c *TransientGetCmd) Run(g *Globals) error {
	return execPassthrough(g, "transient", "get", c.Key)
}

func (c *TransientSetCmd) Run(g *Globals) error {
	parts := []string{"transient", "set", c.Key, c.Value}
	if c.Expiration > 0 {
		parts = append(parts, fmt.Sprint(c.Expiration))
	}
	return execPassthrough(g, parts...)
}

// Ensure TransientSetCmd uses wpcli correctly — the expiration is a positional arg to wp-cli.
var _ = wpcli.New // keep import alive for future use

func (c *TransientDeleteCmd) Run(g *Globals) error {
	if c.All {
		return execPassthrough(g, "transient", "delete", "--all")
	}
	if c.Expired {
		return execPassthrough(g, "transient", "delete", "--expired")
	}
	return execPassthrough(g, "transient", "delete", c.Key)
}

func (c *TransientListCmd) Run(g *Globals) error {
	return execPassthrough(g, "transient", "list")
}
