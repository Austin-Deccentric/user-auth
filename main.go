package main

import (
	"auth/handle"
	"auth/model"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath" // so that we can make path joins compatible on all OS

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var tmpl = template.New("")


func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf8")
	err := tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func main() {
	err := godotenv.Load()
	if err != nil{
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")
	  

	_, err = tmpl.ParseGlob(filepath.Join(".", "templates", "*.html"))
	if err != nil {
		log.Fatalf("Unable to parse templates: %v\n", err)
	}

	fmt.Println(filepath.Join(".", "templates", "*.html"))
	fmt.Println(filepath.Join(".", "templates", "*.css"))

	fs:= http.FileServer(http.Dir("templates/"))
	// Registering routes and handler that we will implement
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/signin", handle.Signin).Methods("POST")
	r.HandleFunc("/signup", handle.Signup).Methods("POST")
	r.HandleFunc("/", handler).Methods("GET")
	r.HandleFunc("/dashboard", handle.Dashboard)
	r.HandleFunc("/refresh", handle.Refresh)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	

	// initialize our database connection
	db := model.InitDB()
	defer db.Close()
	cacheDb := model.InitCache()
	defer cacheDb.Close()
	// start the server on port
	fmt.Printf("Listening and serving on port %s.....", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

