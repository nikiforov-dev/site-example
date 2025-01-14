package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func saveUserToFile(user User) error {
	file, err := os.OpenFile("users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Name: %s, Email: %s\n", user.Name, user.Email))
	return err
}

func handleFormSubmission(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight (OPTIONS) requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = saveUserToFile(user)
	if err != nil {
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "User saved successfully!")
}

func users(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "DELETE" {
		deleteUsers(w, r)
		return
	}

	// Read users from file
	data, err := ioutil.ReadFile("users.txt")
	if err != nil {
		http.Error(w, "Failed to read users file", http.StatusInternalServerError)
		return
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var users []User

	for _, line := range lines {
		parts := strings.SplitN(line, ", ", 2)
		if len(parts) == 2 {
			name := strings.TrimPrefix(parts[0], "Name: ")
			email := strings.TrimPrefix(parts[1], "Email: ")
			users = append(users, User{Name: name, Email: email})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func deleteUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	// Truncate the file
	err := os.Truncate("users.txt", 0)
	if err != nil {
		http.Error(w, "Failed to delete users", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "All users deleted.")
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "index.html")
}

func main() {
	http.HandleFunc("/submit", handleFormSubmission)
	http.HandleFunc("/users", users)
	http.HandleFunc("/", serveHTML)

	fmt.Println("Server is running on http://localhost:1111")
	err := http.ListenAndServe(":1111", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
