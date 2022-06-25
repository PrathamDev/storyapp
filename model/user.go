package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique_index;not null" json:"username"`
	Email    string `gorm:"unique_index;not null" json:"email"`
	Password string `gorm:"not_null" json:"password"`
	PhotoURL string `json:"photourl"`
	Role     int    `json:"role"`
}

func (u *User) IsEmpty() bool {
	if u.Username == "" && u.Email == "" {
		return true
	}
	return false
}
