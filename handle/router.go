package handle

import (
	"auth/model"
	"auth/views"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template" // test with html/template
	"github.com/twinj/uuid"
	"time"
	//"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

const hashCost = 8

// Signup registers an ew user in the database
func Signup(w http.ResponseWriter, r *http.Request){
	// Parse and decode the request body into a new `Credentials` instance
	creds := &views.Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	fmt.Println(creds.Username)
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
		log.Println(err)
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
	fmt.Println("Signing in",creds.Username)
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
			fmt.Println("in here")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// If the error is of any other type, send a 500 status
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Somethin wrong with database",err)
		return
	}
	
	// Compare the stored hashed password, with the hashed version of the password that was received
	if err = bcrypt.CompareHashAndPassword([]byte(hashedCreds.Password), []byte(creds.Password)); err != nil {
		// If the two passwords don't match, return a 401 status
		w.WriteHeader(http.StatusUnauthorized)
		// Render error message on page
		return
	}

	// Create a new random session token
	sessionToken := uuid.NewV4().String()
	// Set the token in the cache, along with the user whom it represents
	// The token has an expiry time of 120 seconds
	_, err = model.Cache.Do("SETEX", sessionToken, "120", creds.Username)
	if err != nil {
		// If there is an error in setting the cache, return an internal server error
		log.Println("Accesing redis seems to have failed",err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Finally, we set the client cookie for "avely_cookie" as the session token we just generated
	// we also set an expiry time of 120 seconds, the same as the cache
	http.SetCookie(w, &http.Cookie{
		Name:    "avely_cookie",
		Value:   sessionToken,
		Expires: time.Now().Add(120 * time.Second),
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{"User Logged In"})

	// render dashboard

	

	// If we reach this point, that means the users password was correct, and that they are authorized

}

// SignupPage for app
func SignupPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ :=template.ParseFiles("./templates/index.html")
	tmpl.Execute(w, nil)
}

// Dashboard authenticates users and renders the dashboard
func Dashboard(w http.ResponseWriter, r *http.Request) {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("avely_cookie")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			// render error page
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		// render error page
		return
	}
	sessionToken := c.Value

	// We then get the name of the user from our cache, where we set the session token
	response, err := model.Cache.Do("GET", sessionToken)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		w.WriteHeader(http.StatusInternalServerError)
		// render error page
		return
	}
	if response == nil {
		// If the session token is not present in cache, return an unauthorized error
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Finally, return the welcome message to the user
	w.Write([]byte(fmt.Sprintf("Welcome %s!", response)))
	// render dashboard page probably using http.redirect
}

// Refresh assigns a new session token and deletes the previous one anytime the usre hits the refresh 
func Refresh(w http.ResponseWriter, r *http.Request) {
	// (BEGIN) The code uptil this point is the same as the first part of the `Welcome` route
	  c, err := r.Cookie("session_token")
	  if err != nil {
		  if err == http.ErrNoCookie {
			  w.WriteHeader(http.StatusUnauthorized)
			  return
		  }
		  w.WriteHeader(http.StatusBadRequest)
		  return
	  }
	  sessionToken := c.Value
  
	  response, err := model.Cache.Do("GET", sessionToken)
	  if err != nil {
		  w.WriteHeader(http.StatusInternalServerError)
		  return
	  }
	  if response == nil {
		  w.WriteHeader(http.StatusUnauthorized)
		  return
	  }
	  // (END) The code uptil this point is the same as the first part of the `Welcome` route
  
	  // Now, create a new session token for the current user
	  newSessionToken := uuid.NewV4().String()
	  _, err = model.Cache.Do("SETEX", newSessionToken, "120", fmt.Sprintf("%s",response))
	  if err != nil {
		  w.WriteHeader(http.StatusInternalServerError)
		  return
	  }
  
	  // Delete the older session token
	  _, err = model.Cache.Do("DEL", sessionToken)
	  if err != nil {
		  w.WriteHeader(http.StatusInternalServerError)
		  return
	  }
  
	  // Set the new token as the users `session_token` cookie
	  http.SetCookie(w, &http.Cookie{
		  Name:    "session_token",
		  Value:   newSessionToken,
		  Expires: time.Now().Add(120 * time.Second),
	  })
  }
