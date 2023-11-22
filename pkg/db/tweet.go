package db

import (
	"github.com/jmoiron/sqlx"

	"gx.com/pkg/models"
)

func AddTweet(db *sqlx.DB, tweet models.NewTweet) (int64, error) {
	var err error

	tx := db.MustBegin()

	res, err := tx.NamedExec(`
    INSERT INTO tweets (user_id, tweet_body, time_posted, likes, retweets, views) 
    VALUES (:user_id, :tweet_body, :time_posted, :likes, :retweets, :views)
    `, tweet)
	if err != nil {
		return -1, err
	}

	err = tx.Commit()
	if err != nil {
		return -1, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return lastId, nil
}

func GetTweetWithUser(db *sqlx.DB, tweetId uint) (models.TweetCollection, error) {
	var err error
	var tweet models.Tweet
	var user models.User

	err = db.Get(&tweet, "SELECT * FROM tweets WHERE tweet_id = $1;", tweetId)
	if err != nil {
		return models.TweetCollection{}, err
	}

	err = db.Get(&user, "SELECT user_id, username, name FROM users WHERE user_id = $1;", tweet.UserId)
	if err != nil {
		return models.TweetCollection{}, err
	}

	return models.TweetCollection{Tweet: tweet, User: user}, nil
}
