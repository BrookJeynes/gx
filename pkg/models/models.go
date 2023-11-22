package models

import (
	"time"

	"github.com/google/uuid"
)

type Tweet struct {
	UserId     uuid.UUID `db:"user_id"`
	TweetId    uint      `db:"tweet_id"`
	TweetBody  string    `db:"tweet_body"`
	TimePosted time.Time `db:"time_posted"`

	Likes    uint `db:"likes"`
	Retweets uint `db:"retweets"`
	Views    uint `db:"views"`
}

type NewTweet struct {
	UserId     uuid.UUID `db:"user_id"`
	TweetBody  string    `db:"tweet_body"`
	TimePosted time.Time `db:"time_posted"`

	Likes    uint `db:"likes"`
	Retweets uint `db:"retweets"`
	Views    uint `db:"views"`
}

type User struct {
	UserId   uuid.UUID `db:"user_id"`
	Username string    `db:"username"`
	Name     string    `db:"name"`
}

type DbUser struct {
	UserId   uuid.UUID `db:"user_id"`
	Username string    `db:"username"`
	Name     string    `db:"name"`
	Password string    `db:"password"`
}

type TweetCollection struct {
	Tweet Tweet
	User  User
}

type IndexState struct {
	User   *User
	Tweets []TweetCollection
}

type ProfileState struct {
	User   User
	Tweets []TweetCollection
}

type Session struct {
	SessionId uuid.UUID `db:"session_id"`
	Data      string    `db:"data"`
	Exp       time.Time `db:"exp"`
}
