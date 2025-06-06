package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/spf13/cobra"
)

var (
	archivePath string
	dryRun      bool
	sortOrder   string
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete tweets",
	Long:  `Deletes tweets based on IDs extracted from Twitter archive file.`,
	Run:   runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVar(&archivePath, "archive", "", "Path to Twitter archive directory")
	deleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Only show tweets that would be deleted without actually deleting")
	deleteCmd.Flags().StringVar(&sortOrder, "sort", "newest", "Sort order: 'newest' (default), 'oldest', or 'original'")
	deleteCmd.MarkFlagRequired("archive")
}

// TwitterArchiveTweet represents a tweet in the Twitter archive JSON format
type TwitterArchiveTweet struct {
	ID        string `json:"id_str"`
	CreatedAt string `json:"created_at"`
	Text      string `json:"full_text"`
	Retweeted bool   `json:"retweeted"`
}

// TwitterAPIResponse represents a response from the Twitter API
type TwitterAPIResponse struct {
	Data struct {
		Deleted bool `json:"deleted"`
	} `json:"data"`
	Errors []struct {
		Title  string `json:"title"`
		Detail string `json:"detail"`
		Type   string `json:"type"`
	} `json:"errors"`
}

// checkTweetExists checks if a tweet still exists using Twitter API
func checkTweetExists(client *http.Client, tweetID string) bool {
	url := fmt.Sprintf("https://api.twitter.com/2/tweets/%s", tweetID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	// Add required headers for Twitter API v2
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error checking tweet %s: %v\n", tweetID, err)
		return false
	}
	defer resp.Body.Close()

	// Parse response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response for tweet %s: %v\n", tweetID, err)
		return false
	}

	// Debug output
	fmt.Printf("Check tweet %s response: HTTP %d - %s\n", tweetID, resp.StatusCode, string(body))

	// 200: tweet exists
	// 404: tweet does not exist (deleted)
	return resp.StatusCode == 200
}

func runDelete(cmd *cobra.Command, args []string) {
	// Check for required credentials
	if consumerKey == "" || consumerSecret == "" || accessToken == "" || accessTokenSecret == "" {
		fmt.Println("Error: All authentication credentials are required")
		return
	}

	// Check that archive path exists
	archiveInfo, err := os.Stat(archivePath)
	if os.IsNotExist(err) {
		fmt.Printf("Error: Archive path %s does not exist\n", archivePath)
		return
	}
	if !archiveInfo.IsDir() {
		fmt.Printf("Error: Archive path %s is not a directory\n", archivePath)
		return
	}

	// Find tweet data files in the archive
	var tweetFiles []string

	// Try data/tweets.js (current format)
	tweetsFile := filepath.Join(archivePath, "data", "tweets.js")
	if _, err := os.Stat(tweetsFile); err == nil {
		tweetFiles = append(tweetFiles, tweetsFile)
	}

	// Try data/tweets directory (newer format)
	tweetDataDir := filepath.Join(archivePath, "data", "tweets")
	if _, err := os.Stat(tweetDataDir); err == nil {
		files, err := ioutil.ReadDir(tweetDataDir)
		if err == nil {
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".js") {
					tweetFiles = append(tweetFiles, filepath.Join(tweetDataDir, file.Name()))
				}
			}
		}
	}

	// Try tweet.js (older format)
	if len(tweetFiles) == 0 {
		oldFormatFile := filepath.Join(archivePath, "data", "tweet.js")
		if _, err := os.Stat(oldFormatFile); err == nil {
			tweetFiles = append(tweetFiles, oldFormatFile)
		}
	}

	if len(tweetFiles) == 0 {
		fmt.Println("Error: Could not find tweet data files in the archive")
		return
	}

	fmt.Printf("Found %d tweet data files in the archive\n", len(tweetFiles))

	// Parse tweet data files and extract tweets
	var tweets []TwitterArchiveTweet

	for _, file := range tweetFiles {
		fmt.Printf("Reading %s...\n", file)

		// Read file content
		content, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", file, err)
			continue
		}

		// Twitter archive files start with a variable assignment like "window.YTD.tweet.part0 = "
		// We need to remove this prefix to get valid JSON
		jsonContent := content
		if bytes.HasPrefix(content, []byte("window.")) {
			parts := bytes.SplitN(content, []byte("= "), 2)
			if len(parts) < 2 {
				fmt.Printf("Error parsing file %s: unexpected format\n", file)
				continue
			}
			jsonContent = parts[1]
		}

		// Parse JSON
		var tweetData []struct {
			Tweet TwitterArchiveTweet `json:"tweet"`
		}

		err = json.Unmarshal(jsonContent, &tweetData)
		if err != nil {
			fmt.Printf("Error parsing JSON from file %s: %v\n", file, err)
			continue
		}

		// Extract tweets
		for _, t := range tweetData {
			tweets = append(tweets, t.Tweet)
		}
	}

	if len(tweets) == 0 {
		fmt.Println("No tweets found in the archive")
		return
	}

	fmt.Printf("Extracted %d tweets from the archive\n", len(tweets))

	// Sort tweets by creation date based on sort order
	switch sortOrder {
	case "oldest":
		sort.Slice(tweets, func(i, j int) bool {
			timeI, _ := time.Parse("Mon Jan 02 15:04:05 -0700 2006", tweets[i].CreatedAt)
			timeJ, _ := time.Parse("Mon Jan 02 15:04:05 -0700 2006", tweets[j].CreatedAt)
			return timeI.Before(timeJ) // oldest first
		})
		fmt.Println("Sorted tweets by date (oldest first)")
	case "newest":
		sort.Slice(tweets, func(i, j int) bool {
			timeI, _ := time.Parse("Mon Jan 02 15:04:05 -0700 2006", tweets[i].CreatedAt)
			timeJ, _ := time.Parse("Mon Jan 02 15:04:05 -0700 2006", tweets[j].CreatedAt)
			return timeI.After(timeJ) // newest first
		})
		fmt.Println("Sorted tweets by date (newest first)")
	default:
		fmt.Printf("Using original order from archive (sort option: %s)\n", sortOrder)
	}

	// Limit to the specified number of tweets
	if limit > 0 {
		// Apply offset first
		if offset > 0 {
			if offset >= len(tweets) {
				fmt.Printf("Offset %d is greater than the number of tweets (%d)\n", offset, len(tweets))
				return
			}
			tweets = tweets[offset:]
			fmt.Printf("Skipped %d tweets as specified by offset\n", offset)
		}

		// Then apply limit
		if len(tweets) > limit {
			tweets = tweets[:limit]
			fmt.Printf("Limited to %d tweets as specified\n", limit)
		}
	}

	// Display tweets
	fmt.Println("\nTweets from your archive:")
	fmt.Println("=========================")

	for i, tweet := range tweets {
		createdAt, _ := time.Parse("Mon Jan 02 15:04:05 -0700 2006", tweet.CreatedAt)
		formattedDate := createdAt.Format("2006/01/02 15:04:05")

		// Truncate text if it's too long
		text := tweet.Text
		if len(text) > 50 {
			text = text[:47] + "..."
		}

		fmt.Printf("%d: [%s] %s (ID: %s)\n", i+1, formattedDate, text, tweet.ID)
	}

	// Confirmation
	if !dryRun {
		fmt.Printf("\nDelete the above %d tweets? [y/N]: ", len(tweets))
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" {
			fmt.Println("Cancelled.")
			return
		}
	} else {
		fmt.Println("\nDry run mode - no tweets will be deleted")
		return
	}

	// Setup OAuth 1.0a authentication
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Execute deletion
	fmt.Println("Checking tweets status and executing deletion...")
	success := 0
	failures := 0
	alreadyDeleted := 0

	// Twitter API v2 only allows 50 requests per 15 minutes for free tier
	const rateLimit = 50
	const rateLimitWindow = 15 * time.Minute

	for i, tweet := range tweets {
		// Check if we need to pause for rate limiting
		if i > 0 && i%rateLimit == 0 {
			fmt.Printf("Rate limit reached. Waiting for %s before continuing...\n", rateLimitWindow)
			time.Sleep(rateLimitWindow)
		}

		// Check if tweet still exists
		if !checkTweetExists(httpClient, tweet.ID) {
			fmt.Printf("Tweet already deleted: ID %s\n", tweet.ID)
			alreadyDeleted++
			continue
		}

		// Create delete request using OAuth 1.0a
		url := fmt.Sprintf("https://api.twitter.com/2/tweets/%s", tweet.ID)
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			fmt.Printf("Error creating request for tweet ID %s: %v\n", tweet.ID, err)
			failures++
			continue
		}

		// Send request with OAuth 1.0a authentication
		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Printf("Error deleting tweet ID %s: %v\n", tweet.ID, err)
			failures++
			continue
		}

		// Parse response
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			fmt.Printf("Failed to delete tweet ID %s: HTTP %d - %s\n", tweet.ID, resp.StatusCode, string(body))

			// If we hit a 429 (Too Many Requests), wait longer
			if resp.StatusCode == 429 {
				fmt.Println("Rate limit exceeded. Waiting for 15 minutes...")
				time.Sleep(15 * time.Minute)
			}

			failures++
		} else {
			success++
			fmt.Printf("Deleted: %d/%d - ID: %s\n", success, len(tweets), tweet.ID)
		}

		// Wait a bit between requests to be nice to the API
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("\nCompleted: Successfully deleted %d/%d tweets. Failed: %d, Already deleted: %d\n",
		success, len(tweets), failures, alreadyDeleted)

	if failures > 0 {
		fmt.Println("\nNote: Some tweets may have failed to delete because:")
		fmt.Println("- They were already deleted")
		fmt.Println("- They were too old (Twitter API limits deletion of older tweets)")
		fmt.Println("- API rate limits were reached")
		fmt.Println("- Your API access level does not allow deletion")
	}

	if success == 0 && failures > 0 {
		fmt.Println("\nAPI Access Restriction Warning:")
		fmt.Println("It appears your Twitter API access level may not allow tweet deletion.")
		fmt.Println("Free tier has very limited access. Consider upgrading to Basic tier ($100/month).")
		fmt.Println("For more information: https://developer.x.com/en/portal/product")
	}
}
