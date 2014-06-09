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

type OAuthClient struct {
	Id          int64
	User        User
	UserId      int64
	RedirectUri string
	Name        string
	ClientId    string
	Secret      string
	AutoConfirm bool
}

type OAuthGrant struct {
	Id            int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Token         string
	OAuthClient   OAuthClient
	OAuthClientId int64
	User          User
	UserId        int64
}

func InitModels() {
	gDb.AutoMigrate(User{})
	gDb.AutoMigrate(UserEmail{})
	gDb.AutoMigrate(UserLogin{})
	gDb.AutoMigrate(OAuthClient{})
	gDb.AutoMigrate(OAuthGrant{})
}
