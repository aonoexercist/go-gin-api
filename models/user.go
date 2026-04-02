package models

type User struct {
	ID       uint    `gorm:"primaryKey" json:"id"`
	Name     string  `json:"name" binding:"required"`
	Email    string  `gorm:"unique" json:"email" binding:"required,email"`
	Password string  `json:"-" binding:"required"`
	GoogleID *string `gorm:"unique" json:"google_id,omitempty"`
	Provider string  `json:"provider,omitempty"`
	Roles    []Role  `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type GoogleAuthRequest struct {
	Token string `json:"token"`
}

type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"unique" json:"name"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

type Permission struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique" json:"name"`
}
