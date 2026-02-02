// FutBoss AI - Team Command
// Author: Bruno Lucena (bruno@lucena.cloud)

package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/brunolucena/futboss-ai/cli/internal/ui"
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Manage your team",
	Long:  `View and manage your team roster, formation, and tactics.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(ui.NewTeamModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(teamCmd)
}

