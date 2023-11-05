package model

import "gorm.io/gorm"

type RoleModel struct {
	DB *gorm.DB
}

type Role struct {
	gorm.Model
	Name string `gorm:"type:varchar(255)"`
}
