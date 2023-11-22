package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"gx.com/pkg/configs"
	"gx.com/pkg/crypto"
	"gx.com/pkg/db"
)

func CheckUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			ctx := context.WithValue(r.Context(), "user_id", nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		decryptedSessionId, err := crypto.Decrypt(cookie.Value, configs.Cfg.SecretKey)
		if err != nil {
			ctx := context.WithValue(r.Context(), "user_id", nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		sessionId, err := uuid.Parse(decryptedSessionId)
		if err != nil {
			ctx := context.WithValue(r.Context(), "user_id", nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		isValid, err := db.IsValidSession(configs.Cfg.Db, sessionId)
		if err != nil {
			ctx := context.WithValue(r.Context(), "user_id", nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if !isValid {
			err = db.DeleteSession(configs.Cfg.Db, sessionId)
			if err != nil {
				log.Fatalln(err)
			}
			ctx := context.WithValue(r.Context(), "user_id", nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		session, err := db.GetSession(configs.Cfg.Db, sessionId)
		if err != nil {
			ctx := context.WithValue(r.Context(), "user_id", nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		userId, err := uuid.Parse(session.Data)
		if err != nil {
			ctx := context.WithValue(r.Context(), "user_id", nil)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
