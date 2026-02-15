package cmd

import "fmt"

// Comment commands

type CommentCmd struct {
	List      CommentListCmd      `cmd:"" help:"List comments"`
	Get       CommentGetCmd       `cmd:"" help:"Get comment"`
	Create    CommentCreateCmd    `cmd:"" help:"Create comment"`
	Update    CommentUpdateCmd    `cmd:"" help:"Update comment"`
	Delete    CommentDeleteCmd    `cmd:"" help:"Delete comment"`
	Approve   CommentApproveCmd   `cmd:"" help:"Approve comment"`
	Unapprove CommentUnapproveCmd `cmd:"" help:"Unapprove comment"`
	Spam      CommentSpamCmd      `cmd:"" help:"Mark as spam"`
	Unspam    CommentUnspamCmd    `cmd:"" help:"Unmark as spam"`
	Trash     CommentTrashCmd     `cmd:"" help:"Trash comment"`
	Untrash   CommentUntrashCmd   `cmd:"" help:"Untrash comment"`
	Count     CommentCountCmd     `cmd:"" help:"Count comments"`
}

type (
	CommentListCmd struct{}
	CommentGetCmd  struct {
		ID int `arg:"" help:"Comment ID"`
	}
)

type (
	CommentCreateCmd struct{}
	CommentUpdateCmd struct {
		ID int `arg:"" help:"Comment ID"`
	}
)

type CommentDeleteCmd struct {
	ID int `arg:"" help:"Comment ID"`
}
type CommentApproveCmd struct {
	ID int `arg:"" help:"Comment ID"`
}
type CommentUnapproveCmd struct {
	ID int `arg:"" help:"Comment ID"`
}
type CommentSpamCmd struct {
	ID int `arg:"" help:"Comment ID"`
}
type CommentUnspamCmd struct {
	ID int `arg:"" help:"Comment ID"`
}
type CommentTrashCmd struct {
	ID int `arg:"" help:"Comment ID"`
}
type CommentUntrashCmd struct {
	ID int `arg:"" help:"Comment ID"`
}
type CommentCountCmd struct{}

func (c *CommentListCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "list")
}

func (c *CommentGetCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "get", fmt.Sprint(c.ID))
}

func (c *CommentCreateCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "create")
}

func (c *CommentUpdateCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "update", fmt.Sprint(c.ID))
}

func (c *CommentDeleteCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "delete", fmt.Sprint(c.ID))
}

func (c *CommentApproveCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "approve", fmt.Sprint(c.ID))
}

func (c *CommentUnapproveCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "unapprove", fmt.Sprint(c.ID))
}

func (c *CommentSpamCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "spam", fmt.Sprint(c.ID))
}

func (c *CommentUnspamCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "unspam", fmt.Sprint(c.ID))
}

func (c *CommentTrashCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "trash", fmt.Sprint(c.ID))
}

func (c *CommentUntrashCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "untrash", fmt.Sprint(c.ID))
}

func (c *CommentCountCmd) Run(g *Globals) error {
	return execPassthrough(g, "comment", "count")
}
