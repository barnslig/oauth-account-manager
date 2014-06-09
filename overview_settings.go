package main

import (
	"log"
	"net/http"
)

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
