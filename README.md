# Go Tweet Cleaner

A CLI tool to delete Twitter (X) tweets using your Twitter archive data. You can delete up to 100 tweets at once with flexible sorting options. It asks for confirmation before deletion to prevent accidental removal.

## Features

- ✅ Read tweets from Twitter archive (bypasses API read limitations)
- ✅ Multiple sorting options: newest first, oldest first, or original order
- ✅ Dry run mode for safe testing
- ✅ Rate limiting compliance (50 deletions per 15 minutes for Free tier)
- ✅ Detailed progress reporting
- ✅ Batch processing with customizable limits

## Installation

```bash
go install github.com/kenta/go-tweet-cleaner@latest
```

Alternatively, you can clone the repository and build it manually:

```bash
git clone https://github.com/kenta/go-tweet-cleaner.git
cd go-tweet-cleaner
go build
```

## Required Credentials

To use this tool, you need a Twitter (X) Developer Account and API credentials for an application created there.

1. Visit the [Twitter Developer Portal](https://developer.twitter.com/en/portal/dashboard) to create an account or log in
2. Create an application and obtain the following information:
   - Consumer Key (API Key)
   - Consumer Secret (API Secret)
   - Access Token
   - Access Token Secret

## Twitter Archive

Due to Twitter API restrictions in the Free tier, this tool requires a Twitter archive file to read your tweets. To get your Twitter archive:

1. Go to your Twitter settings
2. Click on "Your account"
3. Click on "Download an archive of your data"
4. Follow the instructions to request and download your archive
5. Extract the archive ZIP file to a directory

## Usage

### Basic Usage

```bash
# Delete newest tweets first (default)
go-tweet-cleaner delete \
  --consumer-key YOUR_CONSUMER_KEY \
  --consumer-secret YOUR_CONSUMER_SECRET \
  --access-token YOUR_ACCESS_TOKEN \
  --access-token-secret YOUR_ACCESS_TOKEN_SECRET \
  --archive /path/to/twitter-archive \
  --limit 10

# Delete oldest tweets first
go-tweet-cleaner delete \
  --consumer-key YOUR_CONSUMER_KEY \
  --consumer-secret YOUR_CONSUMER_SECRET \
  --access-token YOUR_ACCESS_TOKEN \
  --access-token-secret YOUR_ACCESS_TOKEN_SECRET \
  --archive /path/to/twitter-archive \
  --limit 10 \
  --sort oldest
```

### Safe Testing with Dry Run

```bash
# Preview which tweets would be deleted (recommended first step)
go-tweet-cleaner delete \
  --consumer-key YOUR_CONSUMER_KEY \
  --consumer-secret YOUR_CONSUMER_SECRET \
  --access-token YOUR_ACCESS_TOKEN \
  --access-token-secret YOUR_ACCESS_TOKEN_SECRET \
  --archive /path/to/twitter-archive \
  --limit 10 \
  --dry-run
```

### Local Archive Example

If your Twitter archive is in the current directory:

```bash
# Test with dry run
./go-tweet-cleaner delete \
  --consumer-key YOUR_CONSUMER_KEY \
  --consumer-secret YOUR_CONSUMER_SECRET \
  --access-token YOUR_ACCESS_TOKEN \
  --access-token-secret YOUR_ACCESS_TOKEN_SECRET \
  --archive ./twitter-archive \
  --limit 5 \
  --sort oldest \
  --dry-run

# Execute deletion after confirming
./go-tweet-cleaner delete \
  --consumer-key YOUR_CONSUMER_KEY \
  --consumer-secret YOUR_CONSUMER_SECRET \
  --access-token YOUR_ACCESS_TOKEN \
  --access-token-secret YOUR_ACCESS_TOKEN_SECRET \
  --archive ./twitter-archive \
  --limit 5 \
  --sort oldest
```

### Options

- `--consumer-key`: Twitter API Consumer Key (required)
- `--consumer-secret`: Twitter API Consumer Secret (required)
- `--access-token`: Twitter Access Token (required)
- `--access-token-secret`: Twitter Access Token Secret (required)
- `--archive`: Path to Twitter archive directory (required)
- `--limit`: Number of tweets to process at once (default: 100, maximum: 100)
- `--offset`: Number of tweets to skip before processing (default: 0)
- `--dry-run`: Only show tweets that would be deleted without actually deleting
- `--sort`: Sort order for tweets: 'newest' (default), 'oldest', or 'original'

## Execution Flow

1. The tool reads your Twitter archive and extracts tweets
2. Tweets are sorted according to the specified sort order
3. The specified number of tweets (up to the limit) are displayed
4. You will be asked for confirmation to delete; enter `y` to execute the deletion
5. The deletion progress is shown in real-time with rate limiting
6. Upon completion, the number of deleted tweets is displayed

## Sort Options

- **`newest`** (default): Delete most recent tweets first
- **`oldest`**: Delete oldest tweets first (useful for cleaning up old content)
- **`original`**: Use the original order from the Twitter archive

## Notes

- Twitter API Free tier allows only 50 writes per 15 minutes
- The tool will automatically pause when rate limits are reached
- Deleted tweets cannot be restored, so please check carefully before deletion
- Some tweets may fail to delete if they're too old or already deleted
- Twitter archive is used to read tweets due to Twitter API Free tier limitations
- Always use `--dry-run` first to preview what will be deleted

## Twitter API Free Tier Limitations

- 500 Posts per month (posting limit at app level, user level)
- 100 reads per month
- 1 Project, 1 App per Project
- Limited to write-only use cases and media upload endpoints

## Troubleshooting

### Common Issues

1. **"Could not find tweet data files in the archive"**
   - Ensure your archive is properly extracted
   - Check that the archive contains a `data/tweets.js` file

2. **"API Access Restriction Warning"**
   - Your API access level may not allow tweet deletion
   - Consider upgrading to Basic tier ($100/month) for full access

3. **Rate limit errors**
   - The tool automatically handles rate limits
   - Free tier is limited to 50 deletions per 15 minutes

## References
- [Twitter Developer Portal](https://developer.x.com/en/docs/x-api)

## License

MIT