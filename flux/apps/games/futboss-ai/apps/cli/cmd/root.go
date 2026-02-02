// FutBoss AI - CLI Commands
// Author: Bruno Lucena (bruno@lucena.cloud)

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "futboss",
	Short: "FutBoss AI - Football Management Game with AI Agents",
	Long: `FutBoss AI is a multiplayer football management game 
where you manage your team, buy and sell players using tokens,
and compete against other managers powered by AI agents.

The AI agents run locally on Ollama for maximum privacy.

Developer: Bruno Lucena (bruno@lucena.cloud)`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.futboss.yaml)")
	rootCmd.PersistentFlags().String("api-url", "http://localhost:8000", "API server URL")
	rootCmd.PersistentFlags().String("ollama-url", "http://localhost:11434", "Ollama server URL")

	viper.BindPFlag("api_url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("ollama_url", rootCmd.PersistentFlags().Lookup("ollama-url"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".futboss")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}

