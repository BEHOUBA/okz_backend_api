package main

import (
	"log"
	"net/http"

	"./models"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	// serve static images
	filesHandler := http.FileServer(http.Dir("./public"))
	router.PathPrefix("/okz/").Handler(http.StripPrefix("/okz", filesHandler))

	router.HandleFunc("/api/create", models.CreateNewAd).Methods("POST")
	router.HandleFunc("/api/new_ad", models.CreateNewAdID).Methods("GET")
	router.HandleFunc("/api/image/ad/{id}", models.StoreNewImage).Methods("POST")
	router.HandleFunc("/api/register", models.RegisterNewUser).Methods("POST")
	router.HandleFunc("/api/login", models.Login).Methods("POST")
	router.HandleFunc("/api/check", models.CheckUserStatus).Methods("GET")
	router.HandleFunc("/api/get_ads/", models.GetAds).Methods("GET")
	router.HandleFunc("/api/advert/{uid}", models.GetAdByUID).Methods("GET")
	router.HandleFunc("/api/advert_by_categories/", models.AdvertByCategoriesAndCount).Methods("GET")
	router.HandleFunc("/api/get_cities/", models.GetCitiesAndCategoriesInDB).Methods("GET")

	// setup request allowed by the server
	headers := handlers.AllowedHeaders([]string{"Access-Control-Allow-Origin", "Content-Type", "Authorization", "X-Requested-With"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT"})
	origins := handlers.AllowedOrigins([]string{"http://localhost:8080"})

	http.ListenAndServe(getPort(), handlers.CORS(origins, methods, headers)(router))
}

func getPort() string {
	log.Println("server listening on localhost:8008/...")
	return ":8008"
}
