package main

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/justinas/nosurf"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func oAuthAuthorization(w http.ResponseWriter, r *http.Request) {
	var (
		client       oAuthClient
		processError bool
		user         User
	)
	query := r.URL.Query()
	session, _ := SessionStore.Get(r, "user")

	response_type := query.Get("response_type") // 3.1.1, should be code (4.1.1) or token (4.2.1)
	client_id := query.Get("client_id")         // 2.2

	redirect_uri, _ := url.Parse(query.Get("redirect_uri"))
	redirect_uri_split := strings.Split(query.Get("redirect_uri"), "?") // 3.1.2.3
	redirect_query := redirect_uri.Query()

	state := query.Get("state") // 10.12
	if len(state) > 0 {
		redirect_query.Add("state", state)
	}

	if matchUrl := regexp.MustCompile(`https:\/\/(www\.)?`); matchUrl.MatchString(redirect_uri_split[0]) != true {
		// broken redirect uri
		processError = true
		session.AddFlash("Please contact the service owner about this error: Broken redirect_uri")

	} else if response_type != "code" {
		// wrong response_type
		redirect_query.Add("error", "invalid_request")
		redirect_query.Add("error_description", "Wrong response_type!")
		http.Redirect(w, r, fmt.Sprintf("%s?%s", redirect_uri_split[0], redirect_query.Encode()), http.StatusFound)
		return

	} else if err := gDb.Where(&oAuthClient{ClientId: client_id}).First(&client).Error; err != nil {
		// wrong client id
		redirect_query.Add("error", "unauthorized_client")
		redirect_query.Add("error_description", "Unknown client ID!")
		http.Redirect(w, r, fmt.Sprintf("%s?%s", redirect_uri_split[0], redirect_query.Encode()), http.StatusFound)
		return

	} else if client.RedirectUri != redirect_uri_split[0] {
		// wrong redirect uri
		processError = true
		session.AddFlash("Please contact the service owner about this error: Wrong redirect_uri")

	} else if u, err := IsLoggedIn(session); err != nil {
		// user not logged in
		user = u
		http.Redirect(w, r, fmt.Sprintf("/login?redirect=%s", r.URL.RequestURI()), http.StatusMovedPermanently)
		return
	}

	// token-callback !
	if (r.Method == "POST" && !processError) || (client.AutoConfirm) {
		action := "1"
		if !client.AutoConfirm {
			action = r.PostForm["action"][0]
		}

		if action == "1" {
			token := uuid.New()

			// save a token that expires after 10 minutes
			Red.Do("HSET", token, "client_id", client_id)
			Red.Do("HSET", token, "user_id", user.Id)
			Red.Do("EXPIRE", token, "600")

			// redirect!
			redirect_query.Add("code", token)
			http.Redirect(w, r, fmt.Sprintf("%s?%s", redirect_uri_split[0], redirect_query.Encode()), http.StatusFound)
			return

		} else {
			redirect_query.Add("error", "access_denied")
			redirect_query.Add("error_description", "The resource owner denied the request!")
			http.Redirect(w, r, fmt.Sprintf("%s?%s", redirect_uri_split[0], redirect_query.Encode()), http.StatusFound)
			return
		}
	}

	flashes := session.Flashes()
	session.Save(r, w)

	if err := TmploAuthAuth.Execute(w, map[string]interface{}{
		"Title":         "Authorize",
		"_csrf":         nosurf.Token(r),
		"flashes":       flashes,
		"url":           r.URL.RequestURI(),
		"process_error": processError,
		"client":        client,
	}); err != nil {
		log.Fatal(err)
	}
}

func oAuthToken(w http.ResponseWriter, r *http.Request) {

}
