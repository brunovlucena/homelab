// FutBoss AI - Market Command
// Author: Bruno Lucena (bruno@lucena.cloud)

package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/brunolucena/futboss-ai/cli/internal/ui"
	"github.com/spf13/cobra"
)

var marketCmd = &cobra.Command{
	Use:   "market",
	Short: "Browse the player marketplace",
	Long:  `Browse, buy, and sell players in the transfer market.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(ui.NewMarketModel(), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(marketCmd)
}

