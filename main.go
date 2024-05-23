package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

func main() {
	var err error
	dsn := "root:@tcp(localhost)/recruitment_system"
	db, err = sql.Open("mysql", dsn)
	fmt.Println("DB Connected")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Ping the database to ensure a connection is established
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Create tables if they do not exist
	createTables()

	router := mux.NewRouter()

	router.HandleFunc("/signup", Signup).Methods("POST")
	router.HandleFunc("/login", Login).Methods("POST")
	router.HandleFunc("/uploadResume", AuthMiddleware(UploadResume)).Methods("POST")
	router.HandleFunc("/admin/job", AuthMiddleware(CreateJob)).Methods("POST")
	router.HandleFunc("/admin/job/{job_id}", AuthMiddleware(GetJob)).Methods("GET")
	router.HandleFunc("/admin/applicants", AuthMiddleware(GetApplicants)).Methods("GET")
	router.HandleFunc("/admin/applicant/{applicant_id}", AuthMiddleware(GetApplicant)).Methods("GET")
	router.HandleFunc("/jobs", AuthMiddleware(GetJobs)).Methods("GET")
	router.HandleFunc("/jobs/apply", AuthMiddleware(ApplyJob)).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            created_at DATETIME,
            updated_at DATETIME,
            deleted_at DATETIME,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            address TEXT,
            user_type ENUM('Admin', 'Applicant') NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            profile_headline TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS profiles (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            created_at DATETIME,
            updated_at DATETIME,
            deleted_at DATETIME,
            user_id BIGINT NOT NULL,
            resume_file_address TEXT,
            skills TEXT,
            education TEXT,
            experience TEXT,
            phone VARCHAR(20),
            FOREIGN KEY (user_id) REFERENCES users(id)
        )`,
		`CREATE TABLE IF NOT EXISTS jobs (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            created_at DATETIME,
            updated_at DATETIME,
            deleted_at DATETIME,
            title VARCHAR(255) NOT NULL,
            description TEXT,
            posted_on DATETIME,
            total_applications INT DEFAULT 0,
            company_name VARCHAR(255),
            posted_by BIGINT,
            FOREIGN KEY (posted_by) REFERENCES users(id)
        )`,
		`CREATE TABLE IF NOT EXISTS job_applications (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            created_at DATETIME,
            updated_at DATETIME,
            user_id BIGINT NOT NULL,
            job_id BIGINT NOT NULL,
            FOREIGN KEY (user_id) REFERENCES users(id),
            FOREIGN KEY (job_id) REFERENCES jobs(id)
        )`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("Failed to execute query: %s, error: %v", query, err)
		}
	}
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.Split(authHeader, "Bearer ")[1]
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate JWT token and extract user information
		email, err := validateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		r.Header.Set("email", email)
		next(w, r)
	}
}

func validateToken(tokenString string) (string, error) {
	// Parse the token string
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method used
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		// Return the secret key for validation
		return jwtKey, nil
	})

	// Check for parsing errors
	if err != nil {
		return "", fmt.Errorf("error parsing token: %w", err)
	}

	// Validate token claims
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	// Extract user email from claims (assuming "email" claim exists)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", fmt.Errorf("missing email claim in token")
	}

	// Return extracted email and nil error if valid
	return email, nil
}
