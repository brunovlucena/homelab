// FutBoss AI - Match Command
// Author: Bruno Lucena (bruno@lucena.cloud)

package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/brunolucena/futboss-ai/cli/internal/ui"
	"github.com/spf13/cobra"
)

var matchCmd = &cobra.Command{
	Use:   "match",
	Short: "Start a new match",
	Long:  `Start a new match against another player or AI opponent.`,
	Run: func(cmd *cobra.Command, args []string) {
		opponent, _ := cmd.Flags().GetString("opponent")
		p := tea.NewProgram(ui.NewMatchModel(opponent), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running match: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(matchCmd)
	matchCmd.Flags().StringP("opponent", "o", "ai", "Opponent type: 'ai' or player username")
}

