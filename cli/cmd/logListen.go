/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// logListenCmd represents the logListen command
var logListenCmd = &cobra.Command{
	Use:   "log-listen",
	Short: "A command for listen any log files",
	Long:  `Listen any log files and send their to Lognitor app.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("logListen called")
	},
}

func init() {
	rootCmd.AddCommand(logListenCmd)

	logListenCmd.PersistentFlags().String("separator", "\n", "log separator in listened files")
}
