package main

import (
	"time"
)

type UserLogin struct {
	Id        int64
	UserId    int64
	CreatedAt time.Time
}

type UserEmail struct {
	Id         int64
	UserId     int64
	Email      string
	CreatedAt  time.Time
	ActivateId string
	Active     bool
}

type User struct {
	Id        int64
	Realname  string
	Username  string
	Password  string
	Email     []UserEmail
	Active    bool
	CreatedAt time.Time
	Logins    []UserLogin
}

func InitModels() {
	gDb.AutoMigrate(User{})
	gDb.AutoMigrate(UserEmail{})
	gDb.AutoMigrate(UserLogin{})
}
