package main

import (
	"github.com/gorilla/schema"
	"github.com/justinas/nosurf"
	"log"
	"net/http"
	"time"
)

type RegisterForm struct {
	Realname        string `schema:"realname"`
	Username        string `schema:"username"`
	Email           string `schema:"email"`
	Password        string `schema:"password"`
	PasswordConfirm string `schema:"password-confirm"`
	CsrfToken       string `schema:"csrf_token"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var (
		errors  []string
		success []string
	)

	if r.Method == "POST" {
		decoder := schema.NewDecoder()
		register := new(RegisterForm)

		if err := r.ParseForm(); err != nil {
			errors = append(errors, err.Error())
		}
		if err := decoder.Decode(register, r.PostForm); err != nil {
			errors = append(errors, err.Error())
		}
		if register.Password != register.PasswordConfirm {
			errors = append(errors, "Passwords are not identical")
		}

		// create object
		user := User{
			Realname:     register.Realname,
			Username:     register.Username,
			Password:     register.Password,
			RegisteredAt: time.Now(),
			Active:       false,
			Email: []UserEmail{
				{
					Email:     register.Email,
					CreatedAt: time.Now(),
					Active:    false,
				},
			},
		}
		gDb.Save(&user)
		success = append(success, "Registered!")
	}

	if len(errors) != 0 {
		w.WriteHeader(http.StatusBadRequest)
	}
	err := TmplRegister.Execute(w, map[string]interface{}{
		"Title":   "Register",
		"_csrf":   nosurf.Token(r),
		"Errors":  errors,
		"Success": success,
	})
	if err != nil {
		log.Fatal(err)
	}
}
