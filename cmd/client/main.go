package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gophkeeper/internal/client/commands"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	root := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper CLI",
	}

	root.AddCommand(commands.NewVersionCmd(func() (string, string) { return version, buildDate }))
	root.AddCommand(commands.NewRegisterCmd())
	root.AddCommand(commands.NewLoginCmd())
	root.AddCommand(commands.NewAddCmd())
	root.AddCommand(commands.NewListCmd())
	root.AddCommand(commands.NewGetCmd())
	root.AddCommand(commands.NewDeleteCmd())
	root.AddCommand(commands.NewSyncCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
