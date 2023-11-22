package routes

import (
	"log"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	"gx.com/pkg/configs"
	"gx.com/pkg/models"
)

type UsersResource struct{}

func (rs UsersResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/{profile_name}", func(r chi.Router) {
		r.Get("/", rs.Profile)
	})

	return r
}

func (rs UsersResource) Profile(w http.ResponseWriter, r *http.Request) {
	var err error

	database := configs.Cfg.Db
	username := chi.URLParam(r, "profile_name")

	var tweets []models.TweetCollection
	var user models.User

	err = database.Get(&user, "SELECT user_id, username, name FROM users WHERE username = $1;", username)
	if err != nil {
		tmpl, err := template.ParseFiles("templates/404.html")
		if err != nil {
			log.Fatalln(err)
		}

		tmpl.Execute(w, nil)
        return;
	}

	rows, err := database.Queryx("SELECT * FROM tweets WHERE user_id = $1;", user.UserId)
	var tweet models.Tweet
	for rows.Next() {
		err := rows.StructScan(&tweet)
		if err != nil {
			log.Fatalln(err)
		}

		tweets = append(tweets, models.TweetCollection{Tweet: tweet, User: user})
	}

    tmpl, err := template.ParseFiles("templates/profile.html", "templates/tweet.html")
	if err != nil {
		log.Fatalln(err)
	}

	tmpl.Execute(w, models.ProfileState{Tweets: tweets, User: user})
}
