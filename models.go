package main

import "time"

type SignupRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	UserType        string `json:"user_type"`
	ProfileHeadline string `json:"profile_headline"`
	Address         string `json:"address"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type Job struct {
	ID                int       `json:"id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	PostedOn          time.Time `json:"posted_on"`
	CompanyName       string    `json:"company_name"`
	TotalApplications int       `json:"total_applications"`
}
type CreateJobRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CompanyName string `json:"company_name"`
}
type ResumeDetails struct {
	Skills     string `json:"skills"`
	Education  string `json:"education"`
	Experience string `json:"experience"`
	Phone      string `json:"phone"`
}
type Applicant struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Address         string `json:"address"`
	ProfileHeadline string `json:"profile_headline"`
}
type ApplicantProfile struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Address           string `json:"address"`
	ProfileHeadline   string `json:"profile_headline"`
	ResumeFileAddress string `json:"resume_file_address"`
	Skills            string `json:"skills"`
	Education         string `json:"education"`
	Experience        string `json:"experience"`
	Phone             string `json:"phone"`
}
