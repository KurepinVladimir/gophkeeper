package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"gophkeeper/internal/client/store"
)

func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List local records",
		RunE: func(cmd *cobra.Command, args []string) error {
			rr, err := store.LoadRecords()
			if err != nil {
				return err
			}
			for _, r := range rr {
				fmt.Printf("id=%d type=%s title=%q version=%d deleted=%v updated_at=%s\n",
					r.ID, r.Type, r.Title, r.Version, r.Deleted, r.UpdatedAt.Format(time.RFC3339Nano))
			}
			return nil
		},
	}
}
