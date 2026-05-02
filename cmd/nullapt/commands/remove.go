package commands

import (
	"fmt"

	"github.com/nullapt/nullapt/internal/store"
	"github.com/spf13/cobra"
)

func NewRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <skill>",
		Short:   "Uninstall a skill",
		Aliases: []string{"rm", "uninstall"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			s, err := store.New()
			if err != nil {
				return err
			}
			if err := s.Remove(name); err != nil {
				return err
			}
			fmt.Printf("Removed skill %q\n", name)
			return nil
		},
	}
}
