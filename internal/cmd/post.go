package cmd

import (
	"fmt"

	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Post commands

type PostCmd struct {
	List   PostListCmd   `cmd:"" help:"List posts"`
	Create PostCreateCmd `cmd:"" help:"Create a post"`
	Get    PostGetCmd    `cmd:"" help:"Get post details"`
	Update PostUpdateCmd `cmd:"" help:"Update a post"`
	Delete PostDeleteCmd `cmd:"" help:"Delete a post"`
	Exists PostExistsCmd `cmd:"" help:"Check if post exists"`
	Meta   PostMetaCmd   `cmd:"" help:"Post meta management"`
}

type PostListCmd struct {
	PostType string `help:"Post type" name:"post_type"`
}
type PostCreateCmd struct {
	PostTitle string `help:"Post title" name:"post_title"`
}
type PostGetCmd struct {
	ID int `arg:"" help:"Post ID"`
}
type PostUpdateCmd struct {
	ID int `arg:"" help:"Post ID"`
}
type PostDeleteCmd struct {
	ID    int  `arg:"" help:"Post ID"`
	Force bool `help:"Skip trash"`
}
type PostExistsCmd struct {
	ID int `arg:"" help:"Post ID"`
}
type PostMetaCmd struct {
	List   PostMetaListCmd   `cmd:"" help:"List post meta"`
	Get    PostMetaGetCmd    `cmd:"" help:"Get post meta"`
	Update PostMetaUpdateCmd `cmd:"" help:"Update post meta"`
	Delete PostMetaDeleteCmd `cmd:"" help:"Delete post meta"`
}
type PostMetaListCmd struct {
	ID int `arg:"" help:"Post ID"`
}
type PostMetaGetCmd struct {
	ID  int    `arg:"" help:"Post ID"`
	Key string `arg:"" help:"Meta key"`
}
type PostMetaUpdateCmd struct {
	ID    int    `arg:"" help:"Post ID"`
	Key   string `arg:"" help:"Meta key"`
	Value string `arg:"" help:"Meta value"`
}
type PostMetaDeleteCmd struct {
	ID  int    `arg:"" help:"Post ID"`
	Key string `arg:"" help:"Meta key"`
}

func (c *PostListCmd) Run(g *Globals) error {
	return runStructuredListCommand[wpcli.Post](g, "post list", "", func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("post", "list").Format("json")
		if c.PostType != "" {
			builder.Flag("post_type", c.PostType)
		}
		return builder
	})
}

func (c *PostCreateCmd) Run(g *Globals) error {
	return runWPCommand(g, "post create", nil, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("post", "create")
		if c.PostTitle != "" {
			builder.Flag("post_title", c.PostTitle)
		}
		return builder
	})
}

func (c *PostGetCmd) Run(g *Globals) error {
	return runStructuredSingleCommand[wpcli.Post](g, "post get", func(*registry.Site) *wpcli.Command {
		return wpcli.New("post", "get").Arg(fmt.Sprint(c.ID)).Format("json")
	})
}

func (c *PostUpdateCmd) Run(g *Globals) error {
	return execPassthrough(g, "post", "update", fmt.Sprint(c.ID))
}

func (c *PostDeleteCmd) Run(g *Globals) error {
	return runWPCommand(g, "post delete", nil, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("post", "delete").Arg(fmt.Sprint(c.ID))
		if c.Force {
			builder.BoolFlag("force")
		}
		return builder
	})
}

func (c *PostExistsCmd) Run(g *Globals) error {
	return execBoolCheck(g, "post", "exists", fmt.Sprint(c.ID), "Post")
}

func (c *PostMetaListCmd) Run(g *Globals) error {
	return execPassthrough(g, "post", "meta", "list", fmt.Sprint(c.ID))
}

func (c *PostMetaGetCmd) Run(g *Globals) error {
	return execPassthrough(g, "post", "meta", "get", fmt.Sprint(c.ID), c.Key)
}

func (c *PostMetaUpdateCmd) Run(g *Globals) error {
	return execPassthrough(g, "post", "meta", "update", fmt.Sprint(c.ID), c.Key, c.Value)
}

func (c *PostMetaDeleteCmd) Run(g *Globals) error {
	return execPassthrough(g, "post", "meta", "delete", fmt.Sprint(c.ID), c.Key)
}
