# gator

gator is a simple RSS feed aggre**gator** built in Go. It allows users to follow various RSS feeds, browse downloaded posts, and manage subscriptions to those feeds, all from the CLI.

### Prerequisites

Go
PostgreSQL
goose

### Configuration

gator requires a config file, `.gatorconfig.json` in the home directory of the user in the following format:

```json
{
  "db_url": "postgres://username:@localhost:5432/gator",
  "current_user_name": ""
}
```

Replace the value of `"db_url"` with the appropriate connection string for the Postgres database.

### Usage

register <name>: Creates a user and sets them as the current user
login <name>: Sets an existing user as the current user
users: List all registered users
addfeed <name> <url>: Adds a feed and follows it as the current user
feeds: Lists all feeds
follow <url>: Follows an existing feed as the current user
following: Lists all feeds followed by the current user
unfollow <url>: Unfollows a feed as the current user
agg: Ongoing looping function that fetches feed data
browse <int>: Displays the most recent <int> posts that current user follows
**CAUTION** reset: Delete all users and dependent data!

### Future Improvements

- Finish this README! Need to add instructions for database and goose
- Add `setup` command if possible to allow user to not worry about spinning database up
- A number of tiny bug fixes
- Improve output of posts to the user
- Filtering in `browse` to pull certain feeds (by keyword?)
- API?
- Probably plenty others