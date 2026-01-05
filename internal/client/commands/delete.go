package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"gophkeeper/internal/client/store"
)

func NewDeleteCmd() *cobra.Command {
	var id int64
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Mark local record as deleted (tombstone). Run sync to push.",
		RunE: func(cmd *cobra.Command, args []string) error {
			rr, err := store.LoadRecords()
			if err != nil {
				return err
			}
			ok := false
			for i := range rr {
				if rr[i].ID == id {
					rr[i].Deleted = true
					rr[i].Version++
					rr[i].UpdatedAt = time.Now().UTC()
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("not found")
			}
			return store.SaveRecords(rr)
		},
	}
	cmd.Flags().Int64Var(&id, "id", 0, "record id")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
