package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const schema = `
DROP TABLE IF EXISTS tweets;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    user_id string primary key,
    username text,
    name text,
    password text
);

CREATE TABLE tweets (
	tweet_id integer primary key,
	user_id string,
	tweet_body text,
	time_posted DATETIME,
	likes integer,
	retweets integer,
	views integer,
    foreign key (user_id)
       references users (user_id)
);

CREATE TABLE sessions (
    session_id string primary key,
	data string,
    exp DATETIME
);`


func Init() (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", "posts.db")
	if err != nil {
		return nil, err
	}

	db.MustExec(schema)

	return db, nil
}
