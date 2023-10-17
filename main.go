package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type Response struct {
	Message  string `json:"message"`
	Id       int    `json:"id,omitempty"`
	Location string `json:"location,omitempty"`
	Username string `json:"username,omitempty"`
	Type     string `json:"type,omitempty"`
}

type User struct {
	FirstName string
	LastName  string
	Username  string
	Password  string
}

func ConnectToDatabase() (*sql.DB, error) {
	dsn := "root:@tcp(localhost)/golang_db"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// Signin Handler
func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response := Response{
			Message: "Method not allowed",
			Type:    "error",
		}
		sendJSONResponse(w, response, http.StatusOK)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("ParseForm() error: %v", err), http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		response := Response{
			Message: "Username and password are required",
			Type:    "error",
		}
		sendJSONResponse(w, response, http.StatusOK)
		return
	}

	db, err := ConnectToDatabase()
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	authSuccessful, userID := isAuthenticated(username, password, db)
	if authSuccessful {
		response := Response{
			Message:  "Authentication successful",
			Id:       userID,
			Location: "/client",
			Username: username,
			Type:     "success",
		}
		sendJSONResponse(w, response, http.StatusOK)
	} else {
		response := Response{
			Message: "Authentication failed",
			Type:    "error",
		}
		sendJSONResponse(w, response, http.StatusOK)
	}
}

// Signin Handler
func signupHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		response := Response{
			Message: "Method not allowed",
			Type:    "error",
		}
		sendJSONResponse(w, response, http.StatusOK)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("ParseForm() error: %v", err), http.StatusBadRequest)
		return
	}
	fname := r.FormValue("fname")
	lname := r.FormValue("lname")
	username := r.FormValue("username")
	password := r.FormValue("password")
	cpassword := r.FormValue("cpassword")

	if username == "" || password == "" || fname == "" || lname == "" || cpassword == "" {
		response := Response{
			Message: "All fields are required",
			Type:    "error",
		}
		sendJSONResponse(w, response, http.StatusOK)
		return
	} else if len(password) < 6 {
		response := Response{
			Message: "Password must be 6 or more characters",
			Type:    "error",
		}
		sendJSONResponse(w, response, http.StatusOK)
	} else if password != cpassword {
		response := Response{
			Message: "Confirm password did not match",
			Type:    "error",
		}
		sendJSONResponse(w, response, http.StatusOK)
	} else {
		// Prepare the SQL statement
		db, err := ConnectToDatabase()
		if err != nil {
			response := Response{
				Message: "Error connecting to database",
				Type:    "error",
			}
			sendJSONResponse(w, response, http.StatusOK)
			return
		}
		stmt, err := db.Prepare("INSERT INTO users (first_name, last_name, username, password) VALUES (?, ?, ?, ?)")
		if err != nil {
			response := Response{
				Message: "Internal Server Error",
				Type:    "error",
			}
			sendJSONResponse(w, response, http.StatusOK)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(fname, lname, username, password)
		if err != nil {
			response := Response{
				Message: "Internal Server Error",
				Type:    "error",
			}
			sendJSONResponse(w, response, http.StatusOK)
			return
		} else {
			response := Response{
				Message: "User registered successfully",
				Type:    "success",
			}
			sendJSONResponse(w, response, http.StatusOK)

		}

	}

	db, err := ConnectToDatabase()
	if err != nil {
		response := Response{
			Message: "Error connecting to database",
			Type:    "error",
		}
		sendJSONResponse(w, response, http.StatusOK)
		return
	}
	defer db.Close()

}

func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("JSON encoding error: %v", err), http.StatusInternalServerError)
	}
}

func isAuthenticated(username, password string, db *sql.DB) (bool, int) {
	query := "SELECT id FROM users WHERE username = ? AND password = ?"
	var id int
	err := db.QueryRow(query, username, password).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0
		}
		fmt.Println("Error querying the database:", err)
		return false, 0
	}

	return true, id
}

func main() {

	db, err := ConnectToDatabase()
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer db.Close()

	// start routing
	http.Handle("/", http.FileServer(http.Dir("./static/src")))
	http.Handle("/signin", http.HandlerFunc(signinHandler))
	http.Handle("/signup", http.HandlerFunc(signupHandler))

	// end routing

	fmt.Println("Starting server at port 8080...")
	fmt.Println("Open at http://localhost:8080/")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
