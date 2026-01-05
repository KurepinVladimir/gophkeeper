package commands

import (
	"encoding/base64"
	"time"

	"github.com/spf13/cobra"

	"gophkeeper/internal/client/api"
	"gophkeeper/internal/client/store"
)

func NewSyncCmd() *cobra.Command {
	var server string
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync local records with server (push then pull)",
		RunE: func(cmd *cobra.Command, args []string) error {
			sess, err := mustSession()
			if err != nil {
				return err
			}
			if server == "" {
				server = sess.Server
			}
			c := apiClient(server)

			rr, err := store.LoadRecords()
			if err != nil {
				return err
			}

			// PUSH: send all local records (simple strategy)
			for i := range rr {
				rec := api.Record{
					ID:            rr[i].ID,
					Type:          rr[i].Type,
					Title:         rr[i].Title,
					Meta:          rr[i].Meta,
					EncryptedData: b64ToStd(rr[i].CipherB64),
					Version:       rr[i].Version,
					UpdatedAt:     rr[i].UpdatedAt.Format(time.RFC3339Nano),
					Deleted:       rr[i].Deleted,
				}
				newID, err := c.Upsert(sess.Token, rec)
				if err != nil {
					return err
				}
				if rr[i].ID == 0 {
					rr[i].ID = newID
				}
			}
			if err := store.SaveRecords(rr); err != nil {
				return err
			}

			// PULL: list from server and merge
			serverList, err := c.List(sess.Token)
			if err != nil {
				return err
			}
			merged, err := merge(rr, serverList)
			if err != nil {
				return err
			}
			return store.SaveRecords(merged)
		},
	}
	cmd.Flags().StringVar(&server, "server", "", "server base url (optional, uses saved session)")
	return cmd
}

func b64ToStd(cipherB64 string) string {
	// already base64 from crypto.Encrypt (StdEncoding), keep as-is.
	// validate:
	if _, err := base64.StdEncoding.DecodeString(cipherB64); err != nil {
		// if corrupted, still send as-is; server will reject.
	}
	return cipherB64
}

func merge(local []store.LocalRecord, remote []api.Record) ([]store.LocalRecord, error) {
	index := map[int64]int{}
	for i := range local {
		if local[i].ID != 0 {
			index[local[i].ID] = i
		}
	}

	for _, r := range remote {
		rt, err := time.Parse(time.RFC3339Nano, r.UpdatedAt)
		if err != nil {
			rt = time.Now().UTC()
		}
		if i, ok := index[r.ID]; ok {
			// choose newer by version, then updated_at
			if r.Version > local[i].Version || (r.Version == local[i].Version && rt.After(local[i].UpdatedAt)) {
				local[i].Type = r.Type
				local[i].Title = r.Title
				local[i].Meta = r.Meta
				local[i].CipherB64 = r.EncryptedData
				local[i].Version = r.Version
				local[i].UpdatedAt = rt
				local[i].Deleted = r.Deleted
			}
		} else {
			local = append(local, store.LocalRecord{
				ID: r.ID, Type: r.Type, Title: r.Title, Meta: r.Meta, CipherB64: r.EncryptedData,
				Version: r.Version, UpdatedAt: rt, Deleted: r.Deleted,
			})
		}
	}
	return local, nil
}
