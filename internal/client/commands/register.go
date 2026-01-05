package commands

import (
	"github.com/spf13/cobra"
)

func NewRegisterCmd() *cobra.Command {
	var server, login, password string
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return apiClient(server).Register(login, password)
		},
	}
	cmd.Flags().StringVar(&server, "server", "http://127.0.0.1:8080", "server base url")
	cmd.Flags().StringVar(&login, "login", "", "login")
	cmd.Flags().StringVar(&password, "password", "", "password")
	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}
