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
	router.HandleFunc("/api/image/user/", models.StoreUserProfileImage).Methods("POST")
	router.HandleFunc("/api/register", models.RegisterNewUser).Methods("POST")
	router.HandleFunc("/api/add_to_favorites", models.AddToFavorites).Methods("POST")
	router.HandleFunc("/api/remove_favorite", models.RemoveFavorite).Methods("POST")
	router.HandleFunc("/api/send_message", models.MessageReceiver).Methods("POST")
	router.HandleFunc("/api/login", models.Login).Methods("POST")
	router.HandleFunc("/api/fb_google_login", models.FBAndGoogleLogin).Methods("POST")
	router.HandleFunc("/api/check", models.CheckUserStatus).Methods("GET")
	router.HandleFunc("/api/get_ads/", models.GetAds).Methods("GET")
	router.HandleFunc("/api/getFavorites/", models.GetFavorites).Methods("GET")
	router.HandleFunc("/api/get_user_adverts/", models.GetUserAdverts).Methods("GET")
	router.HandleFunc("/api/advert/{uid}", models.GetAdByUID).Methods("GET")
	router.HandleFunc("/api/advert_by_categories/", models.AdvertByCategoriesAndCount).Methods("GET")
	router.HandleFunc("/api/get_cities/", models.GetCitiesAndCategoriesInDB).Methods("GET")
	router.HandleFunc("/api/delete_ad/", models.DeleteAd).Methods("DELETE")
	router.HandleFunc("/api/update_ad/", models.UpdateAd).Methods("PATCH")
	router.HandleFunc("/api/update_profile/", models.UpdateProfile).Methods("PATCH")

	// setup request allowed by the server
	headers := handlers.AllowedHeaders([]string{"Access-Control-Allow-Origin", "Content-Type", "Authorization", "X-Requested-With"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "PATCH"})
	origins := handlers.AllowedOrigins([]string{"http://localhost:8080"})

	http.ListenAndServe(getPort(), handlers.CORS(origins, methods, headers)(router))
}

func getPort() string {
	log.Println("server listening on localhost:8008/...")
	return ":8008"
}
