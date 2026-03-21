package cmd

import (
	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Cron commands

type CronCmd struct {
	Event    CronEventCmd    `cmd:"" help:"Manage cron events"`
	Schedule CronScheduleCmd `cmd:"" help:"Manage cron schedules"`
	Test     CronTestCmd     `cmd:"" help:"Test WP-Cron spawning"`
}

type CronEventCmd struct {
	List       CronEventListCmd       `cmd:"" help:"List events"`
	Run        CronEventRunCmd        `cmd:"" help:"Run an event"`
	Schedule   CronEventScheduleCmd   `cmd:"" help:"Schedule an event"`
	Unschedule CronEventUnscheduleCmd `cmd:"" help:"Unschedule an event"`
	Delete     CronEventDeleteCmd     `cmd:"" help:"Delete an event"`
}

type (
	CronEventListCmd struct{}
	CronEventRunCmd  struct {
		Hook string `arg:"" help:"Event hook"`
	}
)

type CronEventScheduleCmd struct {
	Hook       string `arg:"" help:"Event hook"`
	NextRun    string `arg:"" help:"Next run time"`
	Recurrence string `arg:"" optional:"" help:"Recurrence"`
}
type CronEventUnscheduleCmd struct {
	Hook string `arg:"" help:"Event hook"`
}
type CronEventDeleteCmd struct {
	Hook string `arg:"" help:"Event hook"`
}

type CronScheduleCmd struct {
	List CronScheduleListCmd `cmd:"" help:"List schedules"`
}

type (
	CronScheduleListCmd struct{}
	CronTestCmd         struct{}
)

func (c *CronEventListCmd) Run(g *Globals) error {
	return runStructuredListCommand[wpcli.CronEvent](g, "cron event list", "", func(*registry.Site) *wpcli.Command {
		return wpcli.New("cron", "event", "list").Format("json")
	})
}

func (c *CronEventRunCmd) Run(g *Globals) error {
	return execPassthrough(g, "cron", "event", "run", c.Hook)
}

func (c *CronEventScheduleCmd) Run(g *Globals) error {
	parts := []string{"cron", "event", "schedule", c.Hook, c.NextRun}
	if c.Recurrence != "" {
		parts = append(parts, c.Recurrence)
	}
	return execPassthrough(g, parts...)
}

func (c *CronEventUnscheduleCmd) Run(g *Globals) error {
	return execPassthrough(g, "cron", "event", "unschedule", c.Hook)
}

func (c *CronEventDeleteCmd) Run(g *Globals) error {
	return execPassthrough(g, "cron", "event", "delete", c.Hook)
}

func (c *CronScheduleListCmd) Run(g *Globals) error {
	return runStructuredListCommand[wpcli.CronSchedule](g, "cron schedule list", "", func(*registry.Site) *wpcli.Command {
		return wpcli.New("cron", "schedule", "list").Format("json")
	})
}

func (c *CronTestCmd) Run(g *Globals) error {
	return execPassthrough(g, "cron", "test")
}
