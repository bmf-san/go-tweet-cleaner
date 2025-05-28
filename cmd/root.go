package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
	limit             int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-tweet-cleaner",
	Short: "Tweet deletion CLI",
	Long:  `A CLI tool to delete Twitter (X) tweets.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&consumerKey, "consumer-key", "", "Twitter API Consumer Key")
	rootCmd.PersistentFlags().StringVar(&consumerSecret, "consumer-secret", "", "Twitter API Consumer Secret")
	rootCmd.PersistentFlags().StringVar(&accessToken, "access-token", "", "Twitter Access Token")
	rootCmd.PersistentFlags().StringVar(&accessTokenSecret, "access-token-secret", "", "Twitter Access Token Secret")
	rootCmd.PersistentFlags().IntVar(&limit, "limit", 100, "Number of tweets to process at once (maximum 100)")

	rootCmd.MarkPersistentFlagRequired("consumer-key")
	rootCmd.MarkPersistentFlagRequired("consumer-secret")
	rootCmd.MarkPersistentFlagRequired("access-token")
	rootCmd.MarkPersistentFlagRequired("access-token-secret")
}
