# osTicket CLI (Go)

A command-line interface for interacting with osTicket using the [unofficial osTicket API](https://bmsvieira.gitbook.io/osticket-api/). Written in Go for fast, single-binary deployment.

## Prerequisites

This CLI requires the [osTicket Unofficial API](https://github.com/BMSVieira/osticket-api) to be installed on your osTicket server.

## Installation

### From Source

```bash
# Clone the repository
cd osticket-cli-go

# Build the binary
go build -o osticket ./cmd/osticket

# Optional: Move to PATH
sudo mv osticket /usr/local/bin/
```

### Pre-built Binary

Download the appropriate binary for your platform from the releases page.

## Configuration

The CLI can be configured via environment variables or a config file. Environment variables take precedence.

### Environment Variables

```bash
export OSTICKET_BASE_URL="https://your-osticket.com/ost_wbs/"
export OSTICKET_API_KEY="YOUR_API_KEY"
```

### Config File

```bash
# Set the API URL and key
osticket config set --url https://your-osticket.com/ost_wbs/ --key YOUR_API_KEY

# View current configuration (shows source: env or config)
osticket config show

# Clear configuration
osticket config clear
```

Configuration file is stored in `~/.osticket-cli/config.yaml`

### Priority

1. Environment variables (`OSTICKET_BASE_URL`, `OSTICKET_API_KEY`)
2. Config file (`~/.osticket-cli/config.yaml`)

## Usage

### Tickets

#### Search/Get Tickets

```bash
# Get a specific ticket by ID or ticket number
osticket ticket get 12345
osticket ticket get API123

# Search tickets by user email
osticket ticket search --email user@example.com

# Search tickets by ticket number
osticket ticket search --number API123

# Search tickets by status (0=all, 1=open, 2=resolved, 3=closed)
osticket ticket search --status 1

# Search tickets by date range
osticket ticket search --from 2024-01-01 --to 2024-12-31

# Output as JSON
osticket ticket search --status 0 --json
```

#### Create Tickets

```bash
# Create a new ticket
osticket ticket create \
  --title "Issue with login" \
  --subject "I cannot log into my account" \
  --user-id 5

# With all options
osticket ticket create \
  --title "Urgent: Server down" \
  --subject "The main server is not responding" \
  --user-id 5 \
  --priority 3 \
  --dept 2 \
  --sla 1 \
  --topic 1
```

#### Reply to Tickets

```bash
osticket ticket reply 12345 \
  --body "We are looking into this issue." \
  --staff-id 1
```

#### Close Tickets

```bash
osticket ticket close 12345 \
  --body "Issue has been resolved." \
  --staff-id 1 \
  --username "admin"
```

### Users

```bash
# Get user by ID
osticket user get --id 5

# Get user by email
osticket user get --email user@example.com

# Create a new user
osticket user create \
  --name "John Doe" \
  --email "john@example.com" \
  --password "secretpassword" \
  --phone "555-1234"
```

### System Information

```bash
# List all departments
osticket info departments

# List all help topics
osticket info topics

# List all SLA plans
osticket info sla
```

## Status Codes

| Status ID | Description |
|-----------|-------------|
| 0 | All (for search) |
| 1 | Open |
| 2 | Resolved |
| 3 | Closed |
| 4 | Archived |
| 5 | Deleted |

## Priority Levels

| Priority ID | Description |
|-------------|-------------|
| 1 | Low |
| 2 | Normal |
| 3 | High |
| 4 | Emergency |

## Building for Multiple Platforms

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o osticket-linux-amd64 ./cmd/osticket

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o osticket-linux-arm64 ./cmd/osticket

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o osticket-darwin-amd64 ./cmd/osticket

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o osticket-darwin-arm64 ./cmd/osticket

# Windows
GOOS=windows GOARCH=amd64 go build -o osticket.exe ./cmd/osticket
```

## License

MIT

## Credits

- [osTicket Unofficial API](https://github.com/BMSVieira/osticket-api) by BMSVieira
