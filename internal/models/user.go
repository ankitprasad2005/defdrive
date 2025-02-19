package models

import (
	"defdrive/internal/db"
	"errors"

	"github.com/go-pg/pg/v10"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

func (u *User) Create() error {
	_, err := db.DB.Model(u).Insert()
	return err
}

func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := db.DB.Model(user).Where("username = ?", username).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}
