package main

import (
	"net/http"
	"log"
)

func IsLoggedIn(w http.ResponseWriter, r *http.Request,  urlStr string) User {
	var user User
	session, _ := SessionStore.Get(r, "user")

	// check if session exists
	if len(session.Values) == 0 {
		http.Redirect(w, r, urlStr, http.StatusMovedPermanently)
	}

	// check if database handle exists
	if gDb.Where(&User{Id: session.Values["id"].(int64)}).First(&user).Error != nil {
		http.Redirect(w, r, urlStr, http.StatusMovedPermanently)
	}

	return user
}

func Overview(w http.ResponseWriter, r *http.Request) {
	user := IsLoggedIn(w, r, "/login")

	if err := TmplOverview.Execute(w, map[string]interface{}{
		"Title": "Login",
		"User": user,
	}); err != nil {
		log.Fatal(err)
	}
}