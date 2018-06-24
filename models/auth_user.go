package models

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"golang.org/x/crypto/bcrypt"
)

var signedKey = []byte("okz")

// User struct for each user on the site
type User struct {
	ID                int       `json:"id"`
	DisplayName       string    `json:"userName"`
	Email             string    `json:"email"`
	PhoneNumber       string    `json:"phoneNumber"`
	Password          string    `json:"password"`
	UserProfileImgURL string    `json:"profileImg"`
	CreatedAt         time.Time `json:"date"`
}

// UserData struct for
type UserData struct {
	Info User
	// Rating       int
	// OwnAdverts   []Advert `json:"userOwnAds`
	// FavoritesAds []Advert `json:"favoritesAds"`
}

// type Session struct {
// 	ID        int
// 	Uuid      string
// 	Email     string
// 	UserID    int
// 	CreatedAt time.Time
// }
type Token struct {
	Token    string   `json:"token"`
	UserData UserData `json:"userdata"`
}

func securityCheck(w http.ResponseWriter, r *http.Request) {
	token, err := request.ParseFromRequest(r, request.OAuth2Extractor, func(token *jwt.Token) (interface{}, error) {
		return signedKey, nil
	})

	if err == nil {
		if token.Valid {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			log.Println("token is not valid")
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("token is not valid")
	}
}
func (u *User) getUserData() (userData UserData, err error) {
	stmt1, err := Db.Prepare("SELECT ID, USER_NAME, EMAIL, PHONE_NUMBER FROM USERS WHERE ID=$1")
	if err != nil {
		return
	}
	err = stmt1.QueryRow(u.ID).Scan(&userData.Info.ID, &userData.Info.DisplayName, &userData.Info.Email, &userData.Info.PhoneNumber)
	if err != nil {
		return
	}

	// userData.OwnAdverts, err = getAdvertsByUserID(u.ID)
	// if err != nil {
	// 	return
	// }
	// userData.FavoritesAds, err = getFavoritesByUserID(u.ID)
	// if err != nil {
	// 	return
	// }
	return
}

func (u *User) getToken(w http.ResponseWriter) error {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["exp"] = time.Now().Add(time.Hour * 48).Unix()
	claims["iss"] = "okz website"
	claims["name"] = u.DisplayName
	tokenValue, err := token.SignedString(signedKey)
	if err != nil {
		return err
	}
	userData, err := u.getUserData()
	if err != nil {
		log.Println(err)
		return err
	}
	tokenJSON, err := json.Marshal(Token{Token: tokenValue, UserData: userData})
	if err != nil {
		log.Println(err)
		return err
	}
	w.Write(tokenJSON)
	return nil
}

func CheckUserStatus(w http.ResponseWriter, r *http.Request) {
	securityCheck(w, r)
}

func Login(w http.ResponseWriter, r *http.Request) {
	// r.ParseMultipartForm(10000)

	// formData := r.MultipartForm.Value
	// var newUser User
	// newUser.Email = strings.Join(formData["email"], "")
	// newUser.Password = strings.Join(formData["password"], "")
	// newUser.ID = 5
	// log.Println(newUser)
	var user User
	bs := make([]byte, r.ContentLength)
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	if err := json.Unmarshal(bs, &user); err != nil {
		log.Println(err)
		return
	}
	if err := user.userAuthentification(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user.getToken(w)
	return
}

func (u *User) userAuthentification() (err error) {
	var DBPassword string
	stmt, err := Db.Prepare("SELECT ID, PASSWORD FROM USERS WHERE EMAIL=$1")
	if err != nil {
		return
	}
	if err := stmt.QueryRow(u.Email).Scan(&u.ID, &DBPassword); err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(DBPassword), []byte(u.Password)); err != nil {
		return err
	}
	return
}

func (u *User) createUser(w http.ResponseWriter) (user User, err error) {
	stmt, err := Db.Prepare("INSERT INTO USERS (USER_NAME, EMAIL, PHONE_NUMBER, PASSWORD, CREATED_AT ) VALUES ($1, $2, $3, $4, $5) returning id, user_name, email, phone_number, password, created_at;")
	if err != nil {
		log.Println(err)
		return
	}
	passwordBS, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		log.Println(err, "error lors du cryptage du mot de passe")
	}
	err = stmt.QueryRow(u.DisplayName, u.Email, u.PhoneNumber, string(passwordBS), time.Now()).Scan(&user.ID, &user.DisplayName, &user.Email, &user.PhoneNumber, &user.Password, &user.CreatedAt)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			w.WriteHeader(http.StatusConflict)
			log.Println("le compte exist deja")
			return User{}, errors.New("ce cette addresse email est deja liée a un compte")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}
	log.Println("user registrated and added to database...")
	return
}

func RegisterNewUser(w http.ResponseWriter, r *http.Request) {
	var data User
	bs := make([]byte, r.ContentLength)
	bs, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	err = json.Unmarshal(bs, &data)
	if err != nil {
		log.Println(err)
		return
	}
	newUser, err := data.createUser(w)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(newUser, "success user created!")
	newUser.getToken(w)
}

func AddToFavorites(w http.ResponseWriter, r *http.Request) {
	advertID, err := strconv.Atoi(r.URL.Query()["id"][0])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userID, err := strconv.Atoi(r.URL.Query()["user"][0])
	if err != nil {
		log.Println("cannot convert userID into integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := addFavoriteToDB(userID, advertID); err != nil {
		log.Println("failed to add favorite to database: ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusAccepted)
}

func addFavoriteToDB(userID int, advertID int) (err error) {
	stmt, err := Db.Prepare("INSERT INTO FAVORITES (USER_ID, ADVERT_ID) VALUES ($1, $2);")
	if err != nil {
		return
	}
	res, err := stmt.Exec(userID, advertID)
	if err != nil {
		return
	}
	log.Println("result of add to favorites statement", res)
	return
}

func RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	advertID, err := strconv.Atoi(r.URL.Query()["id"][0])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userID, err := strconv.Atoi(r.URL.Query()["user"][0])
	if err != nil {
		log.Println("cannot convert userID into integer: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := removeFavoriteFromDB(userID, advertID); err != nil {
		log.Println("failed to remove favorite from database: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func removeFavoriteFromDB(userID int, advertID int) (err error) {
	stmt, err := Db.Prepare("DELETE FROM FAVORITES WHERE USER_ID=$1 AND ADVERT_ID=$2;")
	if err != nil {
		return
	}
	res, err := stmt.Exec(userID, advertID)
	if err != nil {
		return
	}
	log.Println("removed from favorites statement", &res)
	return
}

func GetFavorites(w http.ResponseWriter, r *http.Request) {
	var adverts []Advert
	userID, err := strconv.Atoi(r.URL.Query()["userID"][0])
	if err != nil {
		log.Println(err)
		return
	}
	adverts, err = getFavoritesByUserID(userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	bs, err := json.Marshal(adverts)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bs)
}

func GetUserAdverts(w http.ResponseWriter, r *http.Request) {
	var adverts []Advert
	userID, err := strconv.Atoi(r.URL.Query()["userID"][0])
	if err != nil {
		log.Println(err)
		return
	}
	adverts, err = getAdvertsByUserID(userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	bs, err := json.Marshal(adverts)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bs)
}
