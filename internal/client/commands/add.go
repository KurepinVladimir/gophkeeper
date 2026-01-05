package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"gophkeeper/internal/client/crypto"
	"gophkeeper/internal/client/store"
)

type loginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type textData struct {
	Text string `json:"text"`
}
type cardData struct {
	Number string `json:"number"`
	Exp    string `json:"exp"`
	CVV    string `json:"cvv"`
	Holder string `json:"holder"`
}
type binData struct {
	B64 string `json:"b64"`
}

func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add local record (encrypted on client)",
	}

	cmd.AddCommand(newAddLoginCmd())
	cmd.AddCommand(newAddTextCmd())
	cmd.AddCommand(newAddCardCmd())
	cmd.AddCommand(newAddBinCmd())

	return cmd
}

func newAddLoginCmd() *cobra.Command {
	var title, meta, user, pass, master string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Add login/password record",
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := loadKey(master)
			if err != nil {
				return err
			}
			plain, _ := json.Marshal(loginData{Username: user, Password: pass})
			cipherB64, err := crypto.Encrypt(key, plain)
			if err != nil {
				return err
			}
			return appendLocal(store.LocalRecord{
				Type: "login", Title: title, Meta: meta, CipherB64: cipherB64,
				Version: 1, UpdatedAt: time.Now().UTC(), Deleted: false,
			})
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "title")
	cmd.Flags().StringVar(&meta, "meta", "", "meta")
	cmd.Flags().StringVar(&user, "username", "", "username")
	cmd.Flags().StringVar(&pass, "password", "", "password")
	cmd.Flags().StringVar(&master, "master", "", "master password (for encryption)")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	_ = cmd.MarkFlagRequired("master")
	return cmd
}

func newAddTextCmd() *cobra.Command {
	var title, meta, text, master string
	cmd := &cobra.Command{
		Use:   "text",
		Short: "Add text record",
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := loadKey(master)
			if err != nil {
				return err
			}
			plain, _ := json.Marshal(textData{Text: text})
			cipherB64, err := crypto.Encrypt(key, plain)
			if err != nil {
				return err
			}
			return appendLocal(store.LocalRecord{
				Type: "text", Title: title, Meta: meta, CipherB64: cipherB64,
				Version: 1, UpdatedAt: time.Now().UTC(), Deleted: false,
			})
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "title")
	cmd.Flags().StringVar(&meta, "meta", "", "meta")
	cmd.Flags().StringVar(&text, "text", "", "text")
	cmd.Flags().StringVar(&master, "master", "", "master password (for encryption)")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("text")
	_ = cmd.MarkFlagRequired("master")
	return cmd
}

func newAddCardCmd() *cobra.Command {
	var title, meta, number, exp, cvv, holder, master string
	cmd := &cobra.Command{
		Use:   "card",
		Short: "Add bank card record",
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := loadKey(master)
			if err != nil {
				return err
			}
			plain, _ := json.Marshal(cardData{Number: number, Exp: exp, CVV: cvv, Holder: holder})
			cipherB64, err := crypto.Encrypt(key, plain)
			if err != nil {
				return err
			}
			return appendLocal(store.LocalRecord{
				Type: "card", Title: title, Meta: meta, CipherB64: cipherB64,
				Version: 1, UpdatedAt: time.Now().UTC(), Deleted: false,
			})
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "title")
	cmd.Flags().StringVar(&meta, "meta", "", "meta")
	cmd.Flags().StringVar(&number, "number", "", "card number")
	cmd.Flags().StringVar(&exp, "exp", "", "expiry (MM/YY)")
	cmd.Flags().StringVar(&cvv, "cvv", "", "cvv")
	cmd.Flags().StringVar(&holder, "holder", "", "card holder")
	cmd.Flags().StringVar(&master, "master", "", "master password (for encryption)")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("number")
	_ = cmd.MarkFlagRequired("exp")
	_ = cmd.MarkFlagRequired("cvv")
	_ = cmd.MarkFlagRequired("holder")
	_ = cmd.MarkFlagRequired("master")
	return cmd
}

func newAddBinCmd() *cobra.Command {
	var title, meta, b64, master string
	cmd := &cobra.Command{
		Use:   "bin",
		Short: "Add binary record (base64 payload)",
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := loadKey(master)
			if err != nil {
				return err
			}
			plain, _ := json.Marshal(binData{B64: b64})
			cipherB64, err := crypto.Encrypt(key, plain)
			if err != nil {
				return err
			}
			return appendLocal(store.LocalRecord{
				Type: "bin", Title: title, Meta: meta, CipherB64: cipherB64,
				Version: 1, UpdatedAt: time.Now().UTC(), Deleted: false,
			})
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "title")
	cmd.Flags().StringVar(&meta, "meta", "", "meta")
	cmd.Flags().StringVar(&b64, "b64", "", "base64 payload")
	cmd.Flags().StringVar(&master, "master", "", "master password (for encryption)")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("b64")
	_ = cmd.MarkFlagRequired("master")
	return cmd
}

func appendLocal(r store.LocalRecord) error {
	rr, err := store.LoadRecords()
	if err != nil {
		return err
	}
	rr = append(rr, r)
	if err := store.SaveRecords(rr); err != nil {
		return err
	}
	fmt.Println("saved locally (run sync to push)")
	return nil
}
