package db

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"gx.com/pkg/models"
)

func GetUser(db *sqlx.DB, username string, password string) (*models.User, error) {
	var err error
	var user models.DbUser

	err = db.Get(&user, "SELECT user_id, name, username, password FROM users WHERE username = $1;", username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	return &models.User{UserId: user.UserId, Username: user.Name, Name: user.Username}, nil
}

func AddUser(db *sqlx.DB, user models.DbUser) error {
	var err error

	tx := db.MustBegin()

	_, err = tx.NamedExec("INSERT INTO users (user_id, username, name, password) VALUES (:user_id, :username, :name, :password)", user)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
