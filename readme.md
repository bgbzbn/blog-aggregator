# Gator

Gator is a command-line RSS feed aggregator written in Go and backed by PostgreSQL.

It allows users to register, follow RSS feeds, fetch posts from those feeds, and browse the latest posts directly from the terminal.

## Requirements

To run Gator, you need:

- [Go](https://go.dev/) installed
- [PostgreSQL](https://www.postgresql.org/) installed and running

You will also need a PostgreSQL database for Gator.

## Installation

Install the `gator` CLI with Go:

```bash
go install github.com/bgbzbn/blog-aggregator@latest
```

Make sure your Go binary directory is included in your `PATH`.

You can verify the installation with:

```bash
gator
```

> `go run .` is useful during development, but Gator is intended to be installed and run as a compiled binary.

## Configuration

Gator uses a configuration file named:

```text
~/.gatorconfig.json
```

Create the file in your home directory with the following structure:

```json
{
  "db_url": "postgres://USERNAME:PASSWORD@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Replace:

- `USERNAME` with your PostgreSQL username
- `PASSWORD` with your PostgreSQL password
- `gator` with your database name if you chose a different name

For example:

```json
{
  "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

## Running Gator

Once PostgreSQL is running, the database is set up, and the config file has been created, run commands with:

```bash
gator <command> [arguments]
```

During development, you can also use:

```bash
go run . <command> [arguments]
```

## Commands

### Register a user

```bash
gator register <username>
```

Example:

```bash
gator register lane
```

This creates a user and logs in as that user.

### Login

```bash
gator login <username>
```

Example:

```bash
gator login lane
```

### List users

```bash
gator users
```

### Add a feed

```bash
gator addfeed "<feed-name>" "<feed-url>"
```

Example:

```bash
gator addfeed "Hacker News" "https://hnrss.org/newest"
```

The feed is added and the currently logged-in user follows it.

### List feeds

```bash
gator feeds
```

### Follow a feed

```bash
gator follow "<feed-url>"
```

### View followed feeds

```bash
gator following
```

### Unfollow a feed

```bash
gator unfollow "<feed-url>"
```

### Aggregate feeds

The `agg` command continuously fetches posts from feeds.

```bash
gator agg <time-between-requests>
```

For example:

```bash
gator agg 1m
```

This fetches a feed approximately once per minute.

You can use other Go duration values as well:

```bash
gator agg 30s
```

```bash
gator agg 5m
```

Stop the aggregator with `Ctrl+C`.

### Browse posts

Browse posts from feeds followed by the current user:

```bash
gator browse
```

You can optionally provide the number of posts to display:

```bash
gator browse 10
```

## Example

A typical workflow might look like this:

```bash
gator register lane

gator addfeed "Hacker News" "https://hnrss.org/newest"

gator agg 1m
```

After some posts have been collected, stop the aggregator with `Ctrl+C` and browse them:

```bash
gator browse 10
```

## Development

Clone the repository:

```bash
git clone https://github.com/YOUR_GITHUB_USERNAME/YOUR_REPO_NAME.git
cd YOUR_REPO_NAME
```

Run the application during development with:

```bash
go run . <command>
```

Build a binary with:

```bash
go build
```

Or install it directly:

```bash
go install
```

Because Go applications are statically compiled, the resulting `gator` binary can be run without using `go run`.

## Tech Stack

- Go
- PostgreSQL
- SQLC
- Goose