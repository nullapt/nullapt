package commands

import (
	"fmt"
	"strings"

	"github.com/nullapt/nullapt/internal/store"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed skills",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.New()
			if err != nil {
				return err
			}
			skills, err := s.List()
			if err != nil {
				return err
			}
			if len(skills) == 0 {
				fmt.Println("No skills installed. Run `nullapt get <skill>` to install one.")
				return nil
			}
			fmt.Printf("%-24s %-10s %-20s %s\n", "NAME", "VERSION", "AUTHOR", "TOOLS")
			fmt.Println("───────────────────────────────────────────────────────────────────")
			for _, sk := range skills {
				names := make([]string, len(sk.Interface.Tools))
				for i, t := range sk.Interface.Tools {
					names[i] = t.Name
				}
				fmt.Printf("%-24s %-10s %-20s %s\n", sk.Name, sk.Version, sk.Author, strings.Join(names, ", "))
			}
			return nil
		},
	}
}
