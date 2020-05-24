package model

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
)


// The "db" package level variable will hold the reference to our database instance
var db *sql.DB

//InitDB establishes the database connection
func InitDB() *sql.DB {
	err:= godotenv.Load()
	if err!= nil{
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("host")
	port,_ := strconv.Atoi(os.Getenv("dbport"))
	user:= os.Getenv("user")
	password:= os.Getenv("password")
	dbname:= os.Getenv("dbname")
	// Connect to the postgres db
	//you might have to change the connection string to add your database credentials
	psqlinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port , user ,password, dbname)
	db, err = sql.Open("postgres", psqlinfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err) 
	}
	fmt.Println("Connected to database")
	return db
}

var Cache redis.Conn

//InitCache connects to an instance of redis db
func InitCache() redis.Conn{
	// Initialize the redis connection to a redis instance running on your local machine
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Redis not loade",err)
		panic(err)
	}
	// Assign the connection to the package level `cache` variable
	Cache = conn
	fmt.Println("Connected to cache memory")
	return conn
}