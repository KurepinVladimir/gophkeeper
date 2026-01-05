package commands

import (
	"gophkeeper/internal/client/store"

	"github.com/spf13/cobra"
)

func NewLoginCmd() *cobra.Command {
	var server, login, password string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login and save JWT session locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			tok, err := apiClient(server).Login(login, password)
			if err != nil {
				return err
			}
			return store.SaveSession(store.Session{Server: server, Token: tok})
		},
	}
	cmd.Flags().StringVar(&server, "server", "http://127.0.0.1:8080", "server base url")
	cmd.Flags().StringVar(&login, "login", "", "login")
	cmd.Flags().StringVar(&password, "password", "", "password")
	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}
