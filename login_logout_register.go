package main

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/justinas/nosurf"
	"log"
	"net/http"
	"time"
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
	var (
		errors  []string
		success []string
	)

	if r.Method == "POST" {
		decoder := schema.NewDecoder()
		login := new(LoginForm)

		if err := r.ParseForm(); err != nil {
			errors = append(errors, err.Error())
		}
		if err := decoder.Decode(login, r.PostForm); err != nil {
			errors = append(errors, err.Error())
		}

		// look up the user
		var user User
		if gDb.Where(&User{Username: login.Username, Password: login.Password}).First(&user).Error != nil {
			errors = append(errors, "Wrong username and/or password!")
		} else {
			session, _ := SessionStore.Get(r, "user")
			session.Values["id"] = user.Id
			session.Values["realname"] = user.Realname
			session.Save(r, w)
			success = append(success, "Login successfull!")
		}
	}

	err := TmplLogin.Execute(w, map[string]interface{}{
		"Title": "Login",
		"_csrf": nosurf.Token(r),
		"Errors":  errors,
		"Success": success,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {

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
		if err := SendMail(register.Email, "Activate your Account", "/confirm/" + user.Email[0].ActivateId); err != nil {
			errors = append(errors, err.Error())
		} else {
			gDb.Save(&user)
			success = append(success, "Registered!")
		}
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
