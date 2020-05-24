package handle

import (
	"auth/model"
	"auth/views"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"golang.org/x/crypto/bcrypt"
)

const hashCost = 8





// Signup registers an ew user in the database
func Signup(w http.ResponseWriter, r *http.Request){
	// Parse and decode the request body into a new `Credentials` instance
	creds := &views.Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	fmt.Println(creds)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return 
	}
	// Salt and hash the password using the bcrypt algorithm
	// The second argument is the cost of hashing, which we arbitrarily set as 8 (this value can be more or less, depending on the computing power you wish to utilize)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), hashCost)
	username := creds.Username
	passkey := string(hashedPassword)
	// Next, insert the username, along with the hashed , password into the database
	if err = model.CreateUser(username, passkey); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(creds.Username)
	// We reach this point if the credentials we correctly stored in the database, and the default status of 200 is sent back
}

// Signin authenticates user login credentials
func Signin(w http.ResponseWriter, r *http.Request){
	// Parse and decode the request body into a new `Credentials` instance	
	creds := &views.Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status		
		w.WriteHeader(http.StatusBadRequest)
		return 
	}
	
	// We create another instance of `Credentials` to store the credentials we get from the database
	hashedCreds,err := model.GetUserCredential(creds.Username)
	if err != nil {
		// If an entry with the username does not exist, send an "Unauthorized"(401) status
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// If the error is of any other type, send a 500 status
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	

	// Compare the stored hashed password, with the hashed version of the password that was received
	if err = bcrypt.CompareHashAndPassword([]byte(hashedCreds.Password), []byte(creds.Password)); err != nil {
		// If the two passwords don't match, return a 401 status
		w.WriteHeader(http.StatusUnauthorized)
		// Render error message on page
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{"User Logged In"})

	// render dashboard

	

	// If we reach this point, that means the users password was correct, and that they are authorized

}

func SignupPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ :=template.ParseFiles("./templates/index.html")
	tmpl.Execute(w, nil)
}

