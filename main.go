package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/bgbzbn/blog-aggregator/internal/config"
	"github.com/bgbzbn/blog-aggregator/internal/database"
	"github.com/bgbzbn/blog-aggregator/internal/rss"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("username is required")
	}

	username := cmd.args[0]

	userCount, err := s.db.CountUsersByName(context.Background(), username)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if userCount == 0 {
		fmt.Println("username not registered")
		os.Exit(1)
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return err
	}

	fmt.Printf("User set to %s\n", username)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("username is required")
	}

	userID := uuid.New()
	userName := cmd.args[0]

	userCount, err := s.db.CountUsersByName(context.Background(), userName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if userCount > 0 {
		fmt.Println("username already registered")
		os.Exit(1)
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      userName,
	})
	if err != nil {
		return err
	}

	// s.cfg.CurrentUserName = userName
	s.cfg.SetUser(userName)

	// fmt.Println("User:")
	// fmt.Println(user)

	fmt.Printf("Username %s was registered!\n", user.Name)

	return nil
}

func handlerResetUsers(s *state, cmd command) error {
	err := s.db.ResetFeedFollows(context.Background())
	if err != nil {
		return err
	}

	err = s.db.ResetFeeds(context.Background())
	if err != nil {
		return err
	}

	err = s.db.ResetUsers(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func handlerListUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for i := range len(users) {
		user := users[i]
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

func handlerAggregator(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("time duration required  1s, 1m, 1h")
	}

	duration, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for i := range feeds {
		feed := feeds[i]
		fmt.Printf("* %s@%s(%s)\n", feed.Name, feed.UserName.String, feed.Url)
	}

	// fmt.Println(feed)

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("feed name and url are required")
	}

	feedID := uuid.New()
	feedFollowID := uuid.New()
	// user, err := getCurrentUser(s)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        feedID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	followedFeed, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        feedFollowID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Printf("Feed %s(%s) was added!\n", feed.Name, feed.Url)
	fmt.Printf("User %s started following %s.\n", followedFeed.UserName, followedFeed.FeedName)

	return nil
}

func handlerFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("feed url is required")
	}

	feedFollowID := uuid.New()

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Println(err)
		return err
	}

	followedFeed, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        feedFollowID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Printf("User %s started following %s.\n", followedFeed.UserName, followedFeed.FeedName)

	return nil
}

func handlerUserFollows(s *state, cmd command, user database.User) error {
	// user, err := getCurrentUser(s)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }

	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Printf("User %s follows:\n", user.Name)

	for x := range feeds {
		feed := feeds[x]
		fmt.Printf(" - %s\n", feed.FeedName)
	}

	return nil
}

func handlerUnfollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("feed url is required")
	}

	// feedFollowID := uuid.New()

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Printf("User %s unfollowed %s.\n", user.Name, feed.Name)

	return nil
}

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	rssFeed, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Fetched feed %s\n", rssFeed.Channel.Title)

	for i := range rssFeed.Channel.Item {
		item := rssFeed.Channel.Item[i]
		postID := uuid.New()

		publishedAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			fmt.Printf("Failed to parse date %s", item.PubDate)
			continue
		}

		post, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:        postID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title:     item.Title,
			Url:       item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  item.Description != "",
			},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})

		if err != nil {
			fmt.Printf("Failed to save post %s(%s) to the database.\n", item.Title, item.Link)
		}

		fmt.Printf("Saved post %s(%s) to the database.\n", post.Title, post.Url)
	}
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	nrPosts := 2
	if len(cmd.args) == 1 {
		i, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			fmt.Printf("Invalid posts limit number, setting default 2!")
		} else {
			nrPosts = i
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(nrPosts),
	})
	if err != nil {
		return err
	}

	for i := range posts {
		post := posts[i]
		fmt.Printf("* %s - %s\n", post.Title, post.Url)
	}

	return nil
}

// func getCurrentUser(s *state) (database.User, error) {
// 	currentUsername := s.cfg.CurrentUserName
// 	user, err := s.db.GetUserByName(context.Background(), currentUsername)
// 	if err != nil {
// 		return database.User{}, err
// 	}

// 	return user, nil
// }

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unkown command: %s", cmd.name)
	}
	return handler(s, cmd)
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, c command) error {
		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("Could not get current user: %w", err)
		}

		return handler(s, c, user)
	}
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	dbQueries := database.New(db)

	s := &state{
		cfg: &cfg,
		db:  dbQueries,
	}

	cmds := commands{
		handlers: make(map[string]func(*state, command) error),
	}

	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("agg", handlerAggregator)
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))
	cmds.register("feeds", handlerListFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollowFeed))
	cmds.register("following", middlewareLoggedIn(handlerUserFollows))
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerResetUsers)
	cmds.register("users", handlerListUsers)
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollowFeed))

	if len(os.Args) < 2 {
		fmt.Println("not enough arguments")
		os.Exit(1)
	}

	commandName := os.Args[1]
	commandArgs := os.Args[2:]

	cmd := command{
		name: commandName,
		args: commandArgs,
	}

	err = cmds.run(s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
