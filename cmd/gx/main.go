package main

import (
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"gx.com/pkg/configs"
	"gx.com/pkg/crypto"
	"gx.com/pkg/db"
	"gx.com/pkg/middleware"
	"gx.com/pkg/models"
	"gx.com/pkg/routes"

	"html/template"
	"log"
	"net/http"
)

func main() {
	var err error

	err = godotenv.Load()
	if err != nil {
		log.Fatal("[ERROR] Failed to load .env file")
	}

	configs.Cfg.SecretKey = os.Getenv("SECRET_KEY")
	configs.Cfg.Db, err = db.Init()
	if err != nil {
		log.Fatalln("[ERROR] Failed to initialise database.")
	}
	database := configs.Cfg.Db

	r := chi.NewRouter()

	r.Use(chi_middleware.RequestID)
	r.Use(chi_middleware.RealIP)
	r.Use(chi_middleware.Logger)
	r.Use(chi_middleware.Recoverer)
	r.Use(middleware.CheckUserSession)

	r.Mount("/tweets", routes.TweetsResource{}.Routes())
	r.Mount("/users", routes.UsersResource{}.Routes())

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/index.html", "templates/tweet.html")
		if err != nil {
			w.Write([]byte("<span class=\"text-red-500\">Failed to load page, please try again.</span>"))
			return
		}

		var data []models.TweetCollection
		var user models.User

		rows, err := database.Queryx("SELECT tweet_id FROM tweets;")
		var current_id uint
		for rows.Next() {
			err := rows.Scan(&current_id)
			if err != nil {
				w.Write([]byte("<span class=\"text-red-500\">Failed to load tweets :(</span>"))
				return
			}

			tweet, err := db.GetTweetWithUser(database, current_id)
			if err != nil {
				w.Write([]byte("<span class=\"text-red-500\">Failed to load tweets :(</span>"))
				return
			}

			data = append(data, tweet)
		}

		userId := r.Context().Value("user_id")
		if userId == nil {
			tmpl.Execute(w, models.IndexState{Tweets: data, User: nil})
		} else {
			err = database.Get(&user, "SELECT user_id, username, name FROM users WHERE user_id = $1;", userId)
			if err != nil {
				tmpl.Execute(w, models.IndexState{Tweets: data, User: nil})
			}

			tmpl.Execute(w, models.IndexState{Tweets: data, User: &user})
		}
	})

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			w.Write([]byte("<span class=\"text-red-500\">Failed to load page, please try again.</span>"))
			return
		}

		tmpl.Execute(w, nil)
	})

	r.Get("/signup", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/signup.html")
		if err != nil {
			w.Write([]byte("<span class=\"text-red-500\">Failed to load page, please try again.</span>"))
			return
		}

		tmpl.Execute(w, nil)
	})

	r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		clearCookie(w, "session_id", "")
		if err != nil {
			w.Header().Add("HX-Redirect", "/")
			return
		}

		decryptedSessionId, err := crypto.Decrypt(cookie.Value, configs.Cfg.SecretKey)
		if err != nil {
			w.Header().Add("HX-Redirect", "/")
			return
		}

		sessionId, err := uuid.Parse(decryptedSessionId)
		if err != nil {
			w.Header().Add("HX-Redirect", "/")
			return
		}

		err = db.DeleteSession(configs.Cfg.Db, sessionId)
		if err != nil {
			w.Header().Add("HX-Redirect", "/")
			return
		}

		w.Header().Add("HX-Redirect", "/")
	})

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := db.GetUser(database, username, password)
		if err != nil {
			w.Write([]byte("<span class=\"text-red-500\">Invalid username or password</span>"))
			return
		}

		err = login(w, database, user.UserId)
		if err != nil {
			w.Write([]byte("<span class=\"text-red-500\">Failed to login, please try again.</span>"))
			return
		}

		w.Header().Add("HX-Redirect", "/")
	})

	r.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		username := r.FormValue("username")
		hashed, _ := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), 8)
		uuid := uuid.New()

		err := db.AddUser(configs.Cfg.Db, models.DbUser{UserId: uuid, Username: username, Name: name, Password: string(hashed)})
		if err != nil {
			log.Fatalln(err)
		}

		err = login(w, database, uuid)
		if err != nil {
			w.Write([]byte("<span class=\"text-red-500\">Failed to login, please try again.</span>"))
			return
		}

		w.Header().Add("HX-Redirect", "/")
	})

	http.ListenAndServe(":3000", r)
}

func login(w http.ResponseWriter, database *sqlx.DB, userId uuid.UUID) error {
	var encryptedSessionId string
	session, err := db.IsActiveSession(database, userId)
	if session != nil {
		db.RenewUserSession(database, session.SessionId)
		encryptedSessionId, err = crypto.Encrypt(session.SessionId.String(), configs.Cfg.SecretKey)
		if err != nil {
			return err
		}
	} else {
		session, err := db.AddSession(configs.Cfg.Db, userId.String())
		if err != nil {
			return err
		}
		encryptedSessionId, err = crypto.Encrypt(session.SessionId.String(), configs.Cfg.SecretKey)
		if err != nil {
			return err
		}
	}

	setCookie(w, "session_id", encryptedSessionId)

	return nil
}

func clearCookie(w http.ResponseWriter, key string, value string) {
	cookie := &http.Cookie{
		Name:    key,
		Value:   value,
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
}

func setCookie(w http.ResponseWriter, key string, value string) {
	cookie := &http.Cookie{
		Name:  key,
		Value: value,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
}
