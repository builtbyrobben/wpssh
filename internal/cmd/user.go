package cmd

import (
	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// User commands

type UserCmd struct {
	List          UserListCmd          `cmd:"" help:"List users"`
	Create        UserCreateCmd        `cmd:"" help:"Create a user"`
	Delete        UserDeleteCmd        `cmd:"" help:"Delete a user"`
	Get           UserGetCmd           `cmd:"" help:"Get user details"`
	Update        UserUpdateCmd        `cmd:"" help:"Update a user"`
	SetRole       UserSetRoleCmd       `cmd:"" name:"set-role" help:"Set user role"`
	ResetPassword UserResetPasswordCmd `cmd:"" name:"reset-password" help:"Reset password"`
	Exists        UserExistsCmd        `cmd:"" help:"Check if user exists"`
	Meta          UserMetaCmd          `cmd:"" help:"User meta management"`
}

type UserListCmd struct {
	Role string `help:"Filter by role"`
}
type UserCreateCmd struct {
	Login string `arg:"" help:"User login"`
	Email string `arg:"" help:"User email"`
	Role  string `help:"User role" default:"subscriber"`
}
type UserDeleteCmd struct {
	User string `arg:"" help:"User ID or login"`
}
type UserGetCmd struct {
	User string `arg:"" help:"User ID or login"`
}
type UserUpdateCmd struct {
	User string `arg:"" help:"User ID or login"`
}
type UserSetRoleCmd struct {
	User string `arg:"" help:"User ID or login"`
	Role string `arg:"" help:"Role to assign"`
}
type UserResetPasswordCmd struct {
	User string `arg:"" help:"User ID or login"`
}
type UserExistsCmd struct {
	User string `arg:"" help:"User ID or login"`
}
type UserMetaCmd struct {
	List   UserMetaListCmd   `cmd:"" help:"List user meta"`
	Get    UserMetaGetCmd    `cmd:"" help:"Get user meta"`
	Update UserMetaUpdateCmd `cmd:"" help:"Update user meta"`
	Delete UserMetaDeleteCmd `cmd:"" help:"Delete user meta"`
}
type UserMetaListCmd struct {
	User string `arg:"" help:"User ID"`
}
type UserMetaGetCmd struct {
	User string `arg:"" help:"User ID"`
	Key  string `arg:"" help:"Meta key"`
}
type UserMetaUpdateCmd struct {
	User  string `arg:"" help:"User ID"`
	Key   string `arg:"" help:"Meta key"`
	Value string `arg:"" help:"Meta value"`
}
type UserMetaDeleteCmd struct {
	User string `arg:"" help:"User ID"`
	Key  string `arg:"" help:"Meta key"`
}

func (c *UserListCmd) Run(g *Globals) error {
	return runStructuredListCommand[wpcli.User](g, "user list", cache.CategoryUsers, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("user", "list").Format("json")
		if c.Role != "" {
			builder.Flag("role", c.Role)
		}
		return builder
	})
}

func (c *UserCreateCmd) Run(g *Globals) error {
	return runWPCommand(g, "user create", []string{cache.CategoryUsers, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("user", "create").Arg(c.Login).Arg(c.Email)
		if c.Role != "" {
			builder.Flag("role", c.Role)
		}
		return builder
	})
}

func (c *UserDeleteCmd) Run(g *Globals) error {
	return runWPCommand(g, "user delete", []string{cache.CategoryUsers, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("user", "delete").Arg(c.User)
	})
}

func (c *UserGetCmd) Run(g *Globals) error {
	return runStructuredSingleCommand[wpcli.User](g, "user get", func(*registry.Site) *wpcli.Command {
		return wpcli.New("user", "get").Arg(c.User).Format("json")
	})
}

func (c *UserUpdateCmd) Run(g *Globals) error {
	return runWPCommand(g, "user update", []string{cache.CategoryUsers}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("user", "update").Arg(c.User)
	})
}

func (c *UserSetRoleCmd) Run(g *Globals) error {
	return runWPCommand(g, "user set-role", []string{cache.CategoryUsers}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("user", "set-role").Arg(c.User).Arg(c.Role)
	})
}

func (c *UserResetPasswordCmd) Run(g *Globals) error {
	return runWPCommand(g, "user reset-password", []string{cache.CategoryUsers}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("user", "reset-password").Arg(c.User)
	})
}

func (c *UserExistsCmd) Run(g *Globals) error {
	return execBoolCheck(g, "user", "exists", c.User, "User")
}

func (c *UserMetaListCmd) Run(g *Globals) error {
	return execPassthrough(g, "user", "meta", "list", c.User)
}

func (c *UserMetaGetCmd) Run(g *Globals) error {
	return execPassthrough(g, "user", "meta", "get", c.User, c.Key)
}

func (c *UserMetaUpdateCmd) Run(g *Globals) error {
	return execPassthrough(g, "user", "meta", "update", c.User, c.Key, c.Value)
}

func (c *UserMetaDeleteCmd) Run(g *Globals) error {
	return execPassthrough(g, "user", "meta", "delete", c.User, c.Key)
}
