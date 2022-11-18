package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// Booking ...
type Booking struct {
	ID      int    `json:"id"`
	User    string `json:"user"`
	Members int    `json:"members"`
}

// Server ...
type Server struct {
	DB *gorm.DB
}

// Create booking.
func (s Server) createBooking(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var booking Booking
	json.Unmarshal(reqBody, &booking)
	s.DB.Create(&booking)
	log.Println("succeed to create booking")
	json.NewEncoder(w).Encode(booking)
}

// List bookings.
func (s Server) listBookings(w http.ResponseWriter, r *http.Request) {
	bookings := []Booking{}
	s.DB.Find(&bookings)
	log.Println("succeed to list bookings")
	json.NewEncoder(w).Encode(bookings)
}

func (s Server) handleRequests() {
	log.Println("Starting development server at http://127.0.0.1:8888/")
	log.Println("Quit the server with CONTROL-C.")
	// Creates a new instance of a mux router
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/v1/booking", s.createBooking).Methods("POST")
	router.HandleFunc("/v1/bookings", s.listBookings).Methods("GET")
	log.Fatal(http.ListenAndServe(":8888", router))
}

func main() {
	// Connect to Database.
	// https://gorm.io/docs/connecting_to_the_database.html
	db, err := gorm.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	log.Printf("succeed to connect mysql")
	db.AutoMigrate(&Booking{})
	s := Server{DB: db}
	s.handleRequests()
}
