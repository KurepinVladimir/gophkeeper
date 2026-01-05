package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewVersionCmd(get func() (version, buildDate string)) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and build date",
		Run: func(cmd *cobra.Command, args []string) {
			v, d := get()
			fmt.Printf("version=%s\nbuildDate=%s\n", v, d)
		},
	}
}
