package main

import (
	"errors"
	"github.com/gorilla/sessions"
	"reflect"
)

func TestStructStringsLength(m interface{}) bool {
	r := reflect.ValueOf(m).Elem()

	for i := 0; i < r.NumField(); i++ {
		if r.Type().Field(i).Type.Kind() == reflect.String {
			if len(r.Field(i).Interface().(string)) < 1 {
				return false
			}
		}
	}

	return true
}

func IsLoggedIn(session *sessions.Session) (User, error) {
	var user User

	// check if session exists
	if session.Values["id"] == nil {
		return user, errors.New("No session")
	}

	// check if database handle exists
	if gDb.Where(&User{Id: session.Values["id"].(int64), Active: true}).First(&user).Error != nil {
		return user, errors.New("Session invalid")
	}

	return user, nil
}
