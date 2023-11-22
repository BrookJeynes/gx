package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"gx.com/pkg/configs"
	"gx.com/pkg/db"
	"gx.com/pkg/models"

	"net/http"
	"text/template"
	"time"
)

type TweetsResource struct{}

func (rs TweetsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/create", rs.Create)

	return r
}

func (rs TweetsResource) Create(w http.ResponseWriter, r *http.Request) {
	var err error
	database := configs.Cfg.Db

	userId := r.Context().Value("user_id")
	if userId == nil {
		w.Write([]byte("<span class=\"text-red-500\">Failed to create tweet, please try again.</span>"))
		return
	}

	newTweet := models.NewTweet{
		UserId:     userId.(uuid.UUID),
		TweetBody:  r.FormValue("body"),
		TimePosted: time.Now(),
		Likes:      0,
		Retweets:   0,
		Views:      0,
	}

	id, err := db.AddTweet(database, newTweet)
	if err != nil {
		w.Write([]byte("<span class=\"text-red-500\">Failed to create tweet, please try again.</span>"))
		return
	}

	tweet, err := db.GetTweetWithUser(database, uint(id))
	if err != nil {
		w.Header().Add("HX-Refresh", "true")
		return
	}

	tmpl, err := template.ParseFiles("templates/tweet.html")
	if err != nil {
		w.Header().Add("HX-Refresh", "true")
		return
	}

	tmpl.Execute(w, tweet)
}
