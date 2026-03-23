package model

import "time"

type UserModel struct {
	ID           int64
	Login        string
	HashSalt     string
	PasswordHash string
	CreatedAtUTC time.Time
}

type UserApi struct {
	UserID   int32
	Login    string
	Password string
}
