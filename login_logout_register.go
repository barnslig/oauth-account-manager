package main

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/justinas/nosurf"
	"log"
	"net/http"
	"time"
	"fmt"
)

type LoginForm struct {
	Username string `schema:"username"`
	Password string `schema:"password"`
	CsrfToken       string `schema:"csrf_token"`
}

type RegisterForm struct {
	Realname        string `schema:"realname"`
	Username        string `schema:"username"`
	Email           string `schema:"email"`
	Password        string `schema:"password"`
	PasswordConfirm string `schema:"password-confirm"`
	CsrfToken       string `schema:"csrf_token"`
}

func Confirm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["uuid"]

	// get email
	var mail UserEmail
	if gDb.Where(&UserEmail{ActivateId: id}).First(&mail).Error != nil {
		http.Error(w, "Unknown ID", http.StatusBadRequest)
		return
	}

	// get user
	var user User
	gDb.Model(&mail).Related(&user)

	// activate
	if !mail.Active {
		mail.Active = true
		gDb.Save(&mail)
	}
	if !user.Active {
		user.Active = true
		gDb.Save(&user)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "user")
	fail := false
	var user User

	if _, err := IsLoggedIn(session); err == nil {
		http.Redirect(w, r, "/overview", http.StatusMovedPermanently)
	}

	if r.Method == "POST" {
		decoder := schema.NewDecoder()
		login := new(LoginForm)

		if err := r.ParseForm(); err != nil {
			fail = true
			session.AddFlash(err.Error())
		} else if err := decoder.Decode(login, r.PostForm); err != nil {
			fail = true
			session.AddFlash(err.Error())
		} else if gDb.Where(&User{Username: login.Username, Password: login.Password}).First(&user).Error != nil {
			fail = true
			session.AddFlash("Username and/or password wrong!")
		} else if !user.Active {
			fail = true
			session.AddFlash("User isn't activated!")
		}

		if !fail {
			session.Values["id"] = user.Id
			session.Values["realname"] = user.Realname
			session.Save(r, w)
			http.Redirect(w, r, "/overview", http.StatusMovedPermanently)
		}
	}

	flashes := session.Flashes()
	session.Save(r, w)
	err := TmplLogin.Execute(w, map[string]interface{}{
		"Title": "Login",
		"_csrf": nosurf.Token(r),
		"fail": fail,
		"flashes": flashes,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "user")
	delete(session.Values, "id")
	delete(session.Values, "realname")
	session.AddFlash("Successful logged out!")
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}

func Register(w http.ResponseWriter, r *http.Request) {
	session, _ := SessionStore.Get(r, "user")
	fail := false

	if _, err := IsLoggedIn(session); err == nil {
		http.Redirect(w, r, "/overview", http.StatusMovedPermanently)
	}

	if r.Method == "POST" {
		decoder := schema.NewDecoder()
		register := new(RegisterForm)

		if err := r.ParseForm(); err != nil {
			fail = true
			session.AddFlash(err.Error())
		} else if err := decoder.Decode(register, r.PostForm); err != nil {
			fail = true
			session.AddFlash(err.Error())
		} else if register.Password != register.PasswordConfirm {
			fail = true
			session.AddFlash("Passwords are not identical")
		}

		if !fail {
			// create object
			user := User{
				Realname: register.Realname,
				Username: register.Username,
				Password: register.Password,
				Active:   false,
				Email: []UserEmail{
					{
						Email:      register.Email,
						CreatedAt:  time.Now(),
						Active:     false,
						ActivateId: uuid.NewRandom().String(),
					},
				},
			}

			// send mail
			if err := SendMail(register.Email, "Activate your Account", fmt.Sprintf("%s/confirm/%s", Config.BaseDomain, user.Email[0].ActivateId)); err != nil {
				session.AddFlash(err.Error())
			} else {
				gDb.Save(&user)

				session.AddFlash("Registration successful! Please check your mails to activate your account!")
				session.Save(r, w)
				http.Redirect(w, r, "/login", http.StatusMovedPermanently)
			}
		}
	}

	flashes := session.Flashes()
	session.Save(r, w)
	err := TmplRegister.Execute(w, map[string]interface{}{
		"Title":   "Register",
		"_csrf":   nosurf.Token(r),
		"fail": fail,
		"flashes": flashes,
	})
	if err != nil {
		log.Fatal(err)
	}
}
