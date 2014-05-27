package main

import (
	"errors"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
)

func IsLoggedIn(session *sessions.Session) (User, error) {
	var user User

	// check if session exists
	if session.Values["id"] == nil {
		return user, errors.New("No session")
	}

	// check if database handle exists
	if gDb.Where(&User{Id: session.Values["id"].(int64)}).First(&user).Error != nil {
		return user, errors.New("Session invalid")
	}

	return user, nil
}

func Overview(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "user")

	if user, err := IsLoggedIn(session); err != nil {
		session.AddFlash(err.Error())
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	} else {
		flashes := session.Flashes()
		session.Save(r, w)
		if err := TmplOverview.Execute(w, map[string]interface{}{
			"Title":   "Login",
			"User":    user,
			"flashes": flashes,
		}); err != nil {
			log.Fatal(err)
		}
	}
}
