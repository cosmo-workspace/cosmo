package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	keyUserName = "username"
	keyDeadline = "deadline"
)

type Info struct {
	UserName string
	Deadline int64
}

func Set(sess *sessions.Session, i Info) *sessions.Session {
	sess.Values[keyUserName] = i.UserName
	sess.Values[keyDeadline] = i.Deadline
	return sess
}

func Get(sess *sessions.Session) Info {
	i := Info{}
	if val, ok := sess.Values[keyUserName]; ok {
		if userName, ok := val.(string); ok {
			i.UserName = userName
		}
	}
	if val, ok := sess.Values[keyDeadline]; ok {
		if deadline, ok := val.(int64); ok {
			i.Deadline = deadline
		}
	}
	return i
}

func NewStore(hashKey, blockKey []byte, opt *http.Cookie) sessions.Store {
	store := sessions.NewCookieStore(hashKey, blockKey)

	store.Options.HttpOnly = opt.HttpOnly
	store.Options.SameSite = opt.SameSite
	store.Options.Path = opt.Path
	store.Options.MaxAge = opt.MaxAge
	store.Options.Domain = opt.Domain

	store.MaxAge(store.Options.MaxAge)
	return store
}
