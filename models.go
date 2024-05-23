package main

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name            string
	Email           string `gorm:"unique"`
	Address         string
	UserType        string
	PasswordHash    string
	ProfileHeadline string
	Profile         Profile
}

type Profile struct {
	gorm.Model
	UserID            uint
	ResumeFileAddress string
	Skills            string
	Education         string
	Experience        string
	Phone             string
}

type Job struct {
	gorm.Model
	Title             string
	Description       string
	PostedOn          string
	TotalApplications int
	CompanyName       string
	PostedBy          uint
}
