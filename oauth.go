package main

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/justinas/nosurf"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type oAuthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type oAuthAccessTokenGrantUser struct {
	Id       int64  `json:"id"`
	Fullname string `json:"fullname"`
	Username string `json:"username"`
}

type oAuthAccessTokenGrant struct {
	AccessToken string                    `json:"access_token"`
	TokenType   string                    `json:"token_type"`
	User        oAuthAccessTokenGrantUser `json:"user"`
}

func oAuthAuthorization(w http.ResponseWriter, r *http.Request) {
	var (
		client       OAuthClient
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

	} else if err := gDb.Where(&OAuthClient{ClientId: client_id}).First(&client).Error; err != nil {
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
	if (r.Method == "POST" || client.AutoConfirm) && !processError {
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
	if r.Method == "POST" {
		var (
			client OAuthClient
			grant  OAuthGrant
			user   User
		)
		grant_type := r.FormValue("grant_type")
		code := r.FormValue("code")
		redirect_uri := strings.Split(r.FormValue("redirect_uri"), "?")
		client_id := r.FormValue("client_id")
		red_client_id, _ := redis.String(Red.Do("HGET", code, "client_id"))
		red_user_id, _ := redis.Int64(Red.Do("HGET", code, "user_id"))

		if grant_type != "authorization_code" {
			// wrong grant_type
			j, _ := json.Marshal(oAuthError{
				"unsupported_grant_type",
				"Please use 'authorization_code'!",
			})
			http.Error(w, string(j), http.StatusBadRequest)

		} else if exists, err := redis.Bool(Red.Do("EXISTS", code)); !exists && err != nil {
			// code doesn't exist / is expired
			j, _ := json.Marshal(oAuthError{
				"invalid_grant",
				"The authorization code does not exist or is expired!",
			})
			http.Error(w, string(j), http.StatusBadRequest)

		} else if err := gDb.Where(&OAuthClient{ClientId: client_id}).First(&client).Error; err != nil {
			// client doesn't exist
			j, _ := json.Marshal(oAuthError{
				"invalid_client",
				"The client_id is unknown",
			})
			http.Error(w, string(j), http.StatusBadRequest)

		} else if client.ClientId != red_client_id {
			// client id doesn't match the client id that should request a token with this code
			j, _ := json.Marshal(oAuthError{
				"invalid_client",
				"The supplied client_id does not match the client_id of the supplied authorization code!",
			})
			http.Error(w, string(j), http.StatusBadRequest)
		} else if client.RedirectUri != redirect_uri[0] {
			// wrong redirect uri
			j, _ := json.Marshal(oAuthError{
				"invalid_grant",
				"The redirect_uri does not match!",
			})
			http.Error(w, string(j), http.StatusBadRequest)
		} else if err := gDb.Where(&User{Id: red_user_id}).First(&user).Error; err != nil {
			// user for this authorization code doesn't exist
			j, _ := json.Marshal(oAuthError{
				"invalid_grant",
				"The user for this authorization code does not exist!",
			})
			http.Error(w, string(j), http.StatusBadRequest)
		}

		// create or update the OAuthGrant-table
		token := uuid.New()

		gDb.Where(OAuthGrant{
			OAuthClientId: client.Id,
			UserId:        user.Id,
		}).Attrs(OAuthGrant{
			OAuthClientId: client.Id,
			UserId:        user.Id,
			Token:         token,
		}).FirstOrCreate(&grant)

		j, _ := json.Marshal(oAuthAccessTokenGrant{
			token,
			"bearer",
			oAuthAccessTokenGrantUser{
				user.Id,
				user.Realname,
				user.Username,
			},
		})
		w.Write(j)
	}
}
