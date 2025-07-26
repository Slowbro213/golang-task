package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email       string `json:"email" gorm:"type:varchar(200);"`
	Name        string `json:"name" gorm:"type:varchar(200);"`
	Password    string `json:"password" gorm:"type:varchar(200);"`
	OIDCSubject string `gorm:"uniqueIndex"`
	Post        []Post
}

type OIDCClaims struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"preferred_username"`
	Sub      string `json:"sub"`
}
