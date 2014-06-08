package main

import (
	"time"
)

type oAuthClient struct {
	Id          int64
	UserId      int64
	RedirectUri string
	Name        string
	ClientId    string
	Secret      string
	AutoConfirm bool
}

type oAuthCodes struct {
	Id          int64
	CreatedAt   time.Time
	oAuthClient int64
	Code        string
}

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
	gDb.AutoMigrate(oAuthClient{})
	gDb.AutoMigrate(oAuthCodes{})
	gDb.AutoMigrate(User{})
	gDb.AutoMigrate(UserEmail{})
	gDb.AutoMigrate(UserLogin{})
}
