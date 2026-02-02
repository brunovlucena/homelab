// FutBoss AI - Play Command (Main TUI)
// Author: Bruno Lucena (bruno@lucena.cloud)

package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/brunolucena/futboss-ai/cli/internal/ui"
	"github.com/spf13/cobra"
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Start the interactive game interface",
	Long:  `Launch the full terminal UI to manage your team, play matches, and trade players.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(ui.NewMainModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running program: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(playCmd)
}

