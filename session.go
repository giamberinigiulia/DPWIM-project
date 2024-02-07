package main

import (
	"time"

	"github.com/astaxie/session"
)

var globalSessions *session.Manager
var sessions_new = map[string]Session{}

//each session contains the username of the user and the time at which it expires
type Session struct {
	username string
	expiry   time.Time
}

//function to find out if the session has expired
func (s Session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}
