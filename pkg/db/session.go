package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"gx.com/pkg/models"
)

func AddSession(db *sqlx.DB, data string) (models.Session, error) {
	var err error

	session_id := uuid.New()
	exp := time.Now().Add(time.Hour)
	session := models.Session{SessionId: session_id, Data: data, Exp: exp}

	tx := db.MustBegin()

	_, err = tx.NamedExec(`
    INSERT INTO sessions (session_id, data, exp) 
    VALUES (:session_id, :data, :exp)
    `, session)
	if err != nil {
		return session, err
	}

	err = tx.Commit()
	if err != nil {
		return session, err
	}

	return session, nil
}

func IsActiveSession(db *sqlx.DB, user_id uuid.UUID) (*models.Session, error) {
	var session models.Session

	err := db.Get(&session, "SELECT * FROM sessions WHERE user_id = $1;", user_id)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func GetUserSession(db *sqlx.DB, userId uuid.UUID) (models.Session, error) {
	var err error
	var session models.Session

	err = db.Get(&session, "SELECT * FROM sessions WHERE data = $1;", userId)
	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func RenewUserSession(db *sqlx.DB, sessionId uuid.UUID) error {
	var err error
	exp := time.Now().Add(time.Hour)

	_, err = db.Exec("UPDATE sessions SET exp = $1 WHERE session_id = $2;", exp, sessionId)
	if err != nil {
		return err
	}

	return nil
}

func GetSession(db *sqlx.DB, sessionId uuid.UUID) (models.Session, error) {
	var err error
	var session models.Session

	err = db.Get(&session, "SELECT * FROM sessions WHERE session_id = $1;", sessionId)
	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func DeleteSession(db *sqlx.DB, sessionId uuid.UUID) error {
	var err error

	tx := db.MustBegin()

	_, err = tx.Exec("DELETE FROM sessions WHERE session_id = $1;", sessionId.String())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func IsValidSession(db *sqlx.DB, session_id uuid.UUID) (bool, error) {
	var session_exp time.Time

	row := db.QueryRow("SELECT exp FROM sessions WHERE session_id = $1;", session_id)
	err := row.Scan(&session_exp)
	if err != nil {
		return false, err
	}

	if session_exp.After(time.Now()) {
		return true, nil
	}

	return false, nil
}
