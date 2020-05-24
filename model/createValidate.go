package model

import (
	"auth/views"
	"fmt"
	"log"
	"net/http"
)
// CreateUser creates a new user
func CreateUser(user, password string) error{
	// before creating user check if the user name exits
	if _, err := db.Query("INSERT INTO users (username, password) VALUES ($1,$2)", user, password); err != nil {
		return err
	}
	
	return nil
}

// GetUserCredential gets the existing entry present in the database for the given username
func GetUserCredential(username string) (*views.Credentials,error){
	result := db.QueryRow("SELECT password FROM users WHERE username= $1", username)
	// We create another instance of `Credentials` to store the credentials we get from the database
	hashedCreds := &views.Credentials{}

	// Store the obtained password in `storedCreds`
	err := result.Scan(&hashedCreds.Password)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	
	return hashedCreds,nil
}

var users []views.Signup

func NewUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost{
		http.Error(w, "Must be a post request", http.StatusInternalServerError)
		return
	}
	// we can check if correct data is sent here
	// we can also do authentication here
	fmt.Println("Writing user to db")
	newuser := views.Signup{
		Surname: r.FormValue("Surname"), 
		Firstname: r.FormValue("Firstname"), 
		Username: r.FormValue("Username"), 
		Password: r.FormValue("password"),
	}
	users = append(users, newuser)
	fmt.Println(newuser)
	fmt.Fprintf(w, "New User added: %s\n", newuser.Username)
}