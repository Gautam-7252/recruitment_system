package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("secretkey123")

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func Signup(w http.ResponseWriter, r *http.Request) {
	type SignupRequest struct {
		Name            string `json:"name"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		UserType        string `json:"user_type"`
		ProfileHeadline string `json:"profile_headline"`
		Address         string `json:"address"`
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	query := `INSERT INTO users (name, email, address, user_type, password_hash, profile_headline, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`
	_, err = db.Exec(query, req.Name, req.Email, req.Address, req.UserType, string(passwordHash), req.ProfileHeadline)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func Login(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var passwordHash string
	query := `SELECT password_hash FROM users WHERE email = ?`
	err := db.QueryRow(query, req.Email).Scan(&passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(time.Hour * 24) // Token expires in 24 hours
	claims := &Claims{
		Email: req.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	fmt.Println("JWT TOKEN : ", tokenString)
	fmt.Println("Expiration time : ", expirationTime)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the token in the response header
	w.Header().Set("Authorization", "Bearer "+tokenString)
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	w.WriteHeader(http.StatusOK)
}

func UploadResume(w http.ResponseWriter, r *http.Request) {
	userEmail := r.Header.Get("email")

	// Validate user type
	var userType string
	query := `SELECT user_type FROM users WHERE email = ?`
	err := db.QueryRow(query, userEmail).Scan(&userType)
	if err != nil {
		http.Error(w, "Internal Server Error in usertype", http.StatusInternalServerError)
		return
	}

	if userType != "Applicant" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file type
	if !(header.Filename[len(header.Filename)-4:] == ".pdf" || header.Filename[len(header.Filename)-5:] == ".docx") {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// Store the file temporarily
	tempDir := "/tmp"
	err = os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tempFilePath := tempDir + "/" + header.Filename
	out, err := os.Create(tempFilePath)
	if err != nil {
		http.Error(w, "Error in storing temp file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error copy storing file", http.StatusInternalServerError)
		return
	}

	// Upload the file to the third-party API
	parsedDetails, err := uploadAndParseResume(tempFilePath)
	if err != nil {
		http.Error(w, "Failed to process resume", http.StatusInternalServerError)
		return
	}

	// Insert into the database
	query = `INSERT INTO profiles (user_id, resume_file_address, skills, education, experience, phone, created_at, updated_at)
             VALUES ((SELECT id FROM users WHERE email = ?), ?, ?, ?, ?, ?, NOW(), NOW())
             ON DUPLICATE KEY UPDATE resume_file_address = ?, skills = ?, education = ?, experience = ?, phone = ?, updated_at = NOW()`
	_, err = db.Exec(query, userEmail, tempFilePath, parsedDetails.Skills, parsedDetails.Education, parsedDetails.Experience, parsedDetails.Phone, tempFilePath, parsedDetails.Skills, parsedDetails.Education, parsedDetails.Experience, parsedDetails.Phone)
	if err != nil {
		http.Error(w, "Internal Server Error in query exec", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type ResumeDetails struct {
	Skills     string `json:"skills"`
	Education  string `json:"education"`
	Experience string `json:"experience"`
	Phone      string `json:"phone"`
}

func uploadAndParseResume(filePath string) (*ResumeDetails, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a new HTTP request to the third-party API
	req, err := http.NewRequest("POST", "https://api.apilayer.com/resume_parser/upload", file)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("apikey", "gNiXyflsFu3WNYCz1ZCxdWDb7oQg1Nl1")

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp.Body)

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to upload and parse resume, status: %s, body: %s", resp.Status, string(body))
	}

	// Parse the response JSON
	var details ResumeDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

func CreateJob(w http.ResponseWriter, r *http.Request) {
	type CreateJobRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		CompanyName string `json:"company_name"`
	}

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userEmail := r.Header.Get("email")

	query := `INSERT INTO jobs (title, description, posted_on, total_applications, company_name, posted_by, created_at, updated_at)
              VALUES (?, ?, NOW(), 0, ?, (SELECT id FROM users WHERE email = ?), NOW(), NOW())`
	_, err := db.Exec(query, req.Title, req.Description, req.CompanyName, userEmail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]

	type Job struct {
		ID                int       `json:"id"`
		Title             string    `json:"title"`
		Description       string    `json:"description"`
		PostedOn          time.Time `json:"posted_on"`
		CompanyName       string    `json:"company_name"`
		TotalApplications int       `json:"total_applications"`
	}

	var job Job

	query := `SELECT id, title, description, posted_on, company_name, total_applications
              FROM jobs WHERE id = ?`
	err := db.QueryRow(query, jobID).Scan(&job.ID, &job.Title, &job.Description, &job.PostedOn, &job.CompanyName, &job.TotalApplications)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Job not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(job)
}

func GetApplicants(w http.ResponseWriter, r *http.Request) {
	type Applicant struct {
		ID              int    `json:"id"`
		Name            string `json:"name"`
		Email           string `json:"email"`
		Address         string `json:"address"`
		ProfileHeadline string `json:"profile_headline"`
	}

	var applicants []Applicant

	query := `SELECT id, name, email, address, profile_headline FROM users WHERE user_type = 'Applicant'`
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var applicant Applicant
		err := rows.Scan(&applicant.ID, &applicant.Name, &applicant.Email, &applicant.Address, &applicant.ProfileHeadline)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		applicants = append(applicants, applicant)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(applicants)
}

func GetApplicant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicantID := vars["applicant_id"]

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

	var profile ApplicantProfile

	query := `SELECT u.name, u.email, u.address, u.profile_headline, p.resume_file_address, p.skills, p.education, p.experience, p.phone
              FROM users u
              JOIN profiles p ON u.id = p.user_id
              WHERE u.id = ?`
	err := db.QueryRow(query, applicantID).Scan(&profile.Name, &profile.Email, &profile.Address, &profile.ProfileHeadline,
		&profile.ResumeFileAddress, &profile.Skills, &profile.Education, &profile.Experience, &profile.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Applicant not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(profile)
}

func GetJobs(w http.ResponseWriter, r *http.Request) {
	type Job struct {
		ID                int       `json:"id"`
		Title             string    `json:"title"`
		Description       string    `json:"description"`
		PostedOn          time.Time `json:"posted_on"`
		CompanyName       string    `json:"company_name"`
		TotalApplications int       `json:"total_applications"`
	}

	var jobs []Job

	query := `SELECT id, title, description, posted_on, company_name, total_applications FROM jobs`
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var job Job
		err := rows.Scan(&job.ID, &job.Title, &job.Description, &job.PostedOn, &job.CompanyName, &job.TotalApplications)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(jobs)
}

func ApplyJob(w http.ResponseWriter, r *http.Request) {
	userEmail := r.Header.Get("email")

	// Validate user type
	var userType string
	query := `SELECT user_type FROM users WHERE email = ?`
	err := db.QueryRow(query, userEmail).Scan(&userType)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if userType != "Applicant" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	jobID := r.URL.Query().Get("job_id")

	// Insert into the job_applications table
	query = `INSERT INTO job_applications (user_id, job_id, created_at, updated_at)
             VALUES ((SELECT id FROM users WHERE email = ?), ?, NOW(), NOW())`
	_, err = db.Exec(query, userEmail, jobID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Update the total_applications count for the job
	query = `UPDATE jobs SET total_applications = total_applications + 1 WHERE id = ?`
	_, err = db.Exec(query, jobID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
