package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"gophkeeper/internal/client/crypto"
	"gophkeeper/internal/client/store"
)

func NewGetCmd() *cobra.Command {
	var id int64
	var master string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Decrypt and show local record by id",
		RunE: func(cmd *cobra.Command, args []string) error {
			rr, err := store.LoadRecords()
			if err != nil {
				return err
			}
			var found *store.LocalRecord
			for i := range rr {
				if rr[i].ID == id {
					found = &rr[i]
					break
				}
			}
			if found == nil {
				return fmt.Errorf("not found")
			}
			key, err := loadKey(master)
			if err != nil {
				return err
			}
			plain, err := crypto.Decrypt(key, found.CipherB64)
			if err != nil {
				return err
			}
			fmt.Printf("id=%d type=%s title=%q meta=%q deleted=%v\n", found.ID, found.Type, found.Title, found.Meta, found.Deleted)
			fmt.Println(string(plain))
			return nil
		},
	}
	cmd.Flags().Int64Var(&id, "id", 0, "record id")
	cmd.Flags().StringVar(&master, "master", "", "master password (for decryption)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("master")
	return cmd
}
