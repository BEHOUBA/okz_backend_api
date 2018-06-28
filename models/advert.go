package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // importing postgresql driver to deal with database
	"github.com/satori/go.uuid"
)

// Advert struct representing each ad data
type Advert struct {
	ID          int      `json:"id"`
	OwnerID     int      `json:"userID"`
	UID         string   `json:"Uid"`
	Title       string   `json:"title"`
	Category    string   `json:"category"`
	Location    string   `json:"location"`
	Description string   `json:"description"`
	Price       int      `json:"price"`
	ImgURL      []string `json:"imgUrls"`
	Contact     string   `json:"contact"`
	Address     string   `json:"address"`
	CreatedAt   string   `json:"createdAt"`
	IsFavorite  bool     `json:"isFavorite"`
}

var Db *sql.DB
var err error

func init() {
	Db, err = sql.Open("postgres", "port=5000 sslmode=disable user=postgres dbname=okzdb")
	if err != nil {
		log.Println("error occured when trying to open database connection", err)
		return
	}
}

func (a *Advert) storeNewAdToDB() (err error) {
	var newAdID int
	stmt, err := Db.Prepare("INSERT INTO ADVERTS (LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID, ADDRESS) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id;")
	if err != nil {
		log.Println(err)
		return
	}

	imgStmt, err := Db.Prepare("INSERT INTO IMAGES_URL (ADVERT_ID, URL) VALUES ($1, $2)")
	if err != nil {
		log.Println(err)
		return
	}

	err = stmt.QueryRow(a.Location, a.OwnerID, a.Title, a.Description, a.Category, a.Price, a.Contact, time.Now(), a.UID, a.Address).Scan(&newAdID)
	if err != nil {
		log.Println(err)
		return
	}
	for _, url := range a.ImgURL {
		if _, err := imgStmt.Exec(newAdID, url); err != nil {
			log.Println(err)
			return err
		}
		log.Println("image: ", url, " added to database...")
	}
	log.Println("advert added to database...")
	return nil
}

// CreateNewAd new advert from data comming from front end and then add it to database
func CreateNewAd(w http.ResponseWriter, r *http.Request) {
	var ad Advert
	bs := make([]byte, r.ContentLength)

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	err = json.Unmarshal(bs, &ad)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(ad)
	err = ad.storeNewAdToDB()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("advert successfully created...")
}

// CreateNewAdID generate and unique identifier for each new advert
func CreateNewAdID(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(generateUUID()))
}

func generateUUID() string {
	uuid := uuid.Must(uuid.NewV4())
	return uuid.String()
}

// func GetAds(w http.ResponseWriter, r *http.Request) {
// 	query := r.URL.Query()
// 	limitS := query["limit"]
// 	offsetS := query["offset"]
// 	location := strings.Join(query["location"], "")
// 	category := strings.Join(query["category"], "")
// 	input := strings.Join(query["input"], "")
// 	sort := strings.Join(query["sort"], "")

// 	limit, err := strconv.Atoi(limitS[0])
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	offset, err := strconv.Atoi(offsetS[0])
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	log.Println(limit, offset)
// 	log.Println(location, category, input, sort)
// 	imagesURL, err := getAdvertsFromDB(limit, offset)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	bs, err := json.Marshal(imagesURL)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	w.Write(bs)
// }

func GetAdByUID(w http.ResponseWriter, r *http.Request) {
	adUID := mux.Vars(r)["uid"]
	userID, err := strconv.Atoi(r.URL.Query()["userID"][0])
	advert, err := getAdvertFromDBByUID(adUID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err = advert.isFavorite(userID)
	if err != nil {
		log.Println(err)
	}
	log.Println(advert.IsFavorite)
	bs, err := json.Marshal(advert)
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(bs)
}

func (ad *Advert) isFavorite(userID int) (err error) {
	stmt, err := Db.Prepare("SELECT * FROM FAVORITES WHERE USER_ID=$1 AND ADVERT_ID=$2")
	if err != nil {
		log.Println("Error preparing select statement statement from isFavorite method.")
		return
	}
	res, err := stmt.Exec(userID, ad.ID)
	if err != nil {
		return
	}
	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i == 0 {
		ad.IsFavorite = false
		log.Println("this not favorite")
		return nil
	}
	log.Println("this is favorite")
	ad.IsFavorite = true
	return nil
}

func getAdvertFromDBByUID(UID string) (advert Advert, err error) {
	stmt, err := Db.Prepare("SELECT ID, LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID, ADDRESS FROM ADVERTS WHERE AD_UID=$1")
	if err != nil {
		return
	}
	err = stmt.QueryRow(UID).Scan(&advert.ID, &advert.Location, &advert.OwnerID, &advert.Title, &advert.Description, &advert.Category, &advert.Price, &advert.Contact, &advert.CreatedAt, &advert.UID, &advert.Address)
	if err != nil {
		return
	}
	advert.ImgURL, err = getAdvertImagesURL(advert.ID)
	if err != nil {
		log.Println(err, "failed in gettin images's urls")
		return
	}
	return
}

func getAdvertByID(ID int) (advert Advert, err error) {
	stmt, err := Db.Prepare("SELECT ID, LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID, ADDRESS FROM ADVERTS WHERE ID=$1")
	if err != nil {
		return
	}
	err = stmt.QueryRow(ID).Scan(&advert.ID, &advert.Location, &advert.OwnerID, &advert.Title, &advert.Description, &advert.Category, &advert.Price, &advert.Contact, &advert.CreatedAt, &advert.UID, &advert.Address)
	if err != nil {
		return
	}
	advert.ImgURL, err = getAdvertImagesURL(advert.ID)
	if err != nil {
		log.Println(err, "failed in gettin images's urls")
		return
	}
	return
}

func getAdvertsByUserID(ID int) (ads []Advert, err error) {
	stmt, err := Db.Prepare("SELECT ID, LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID, ADDRESS FROM ADVERTS WHERE OWNER_ID=$1 ORDER BY CREATED_AT DESC")
	if err != nil {
		return
	}
	rows, err := stmt.Query(ID)
	if err != nil {
		return
	}
	for rows.Next() {
		var advert Advert
		err = rows.Scan(&advert.ID, &advert.Location, &advert.OwnerID, &advert.Title, &advert.Description, &advert.Category, &advert.Price, &advert.Contact, &advert.CreatedAt, &advert.UID, &advert.Address)
		if err != nil {
			return
		}
		advert.ImgURL, err = getAdvertImagesURL(advert.ID)
		if err != nil {
			log.Println(err, "failed in gettin images's urls")
		}
		ads = append(ads, advert)
	}
	return
}

func getFavoritesByUserID(ID int) (ads []Advert, err error) {
	stmt, err := Db.Prepare("SELECT ADVERT_ID FROM FAVORITES WHERE USER_ID=$1;")
	if err != nil {
		return
	}
	stmt2, err := Db.Prepare("SELECT ID, LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID, ADDRESS FROM ADVERTS WHERE ID=$1")
	if err != nil {
		return
	}
	rows, err := stmt.Query(ID)
	if err != nil {
		return
	}
	for rows.Next() {
		var adID int
		var advert Advert
		err = rows.Scan(&adID)
		if err != nil {
			return
		}
		err = stmt2.QueryRow(adID).Scan(&advert.ID, &advert.Location, &advert.OwnerID, &advert.Title, &advert.Description, &advert.Category, &advert.Price, &advert.Contact, &advert.CreatedAt, &advert.UID, &advert.Address)
		if err != nil {
			return
		}
		advert.ImgURL, err = getAdvertImagesURL(advert.ID)
		if err != nil {
			log.Println(err, "failed in gettin images's urls")
		}
		ads = append(ads, advert)
	}
	return
}

func getAdvertsFromDB(limit, offset int) (ads []Advert, err error) {
	stmt, err := Db.Prepare("SELECT ID, LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID, ADDRESS FROM ADVERTS ORDER BY CREATED_AT LIMIT $1 OFFSET $2")
	if err != nil {
		log.Println(err)
		return
	}
	rows, err := stmt.Query(limit, offset)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for rows.Next() {
		var ad Advert
		err = rows.Scan(&ad.ID, &ad.Location, &ad.OwnerID, &ad.Title, &ad.Description, &ad.Category, &ad.Price, &ad.Contact, &ad.CreatedAt, &ad.UID, &ad.Address)
		if err != nil {
			log.Println(err)
			return
		}
		ad.ImgURL, err = getAdvertImagesURL(ad.ID)
		if err != nil {
			log.Println(err)
			return
		}
		ads = append(ads, ad)
	}
	return
}

func getAdvertImagesURL(id int) (urls []string, err error) {
	stmt, err := Db.Prepare("SELECT URL FROM IMAGES_URL WHERE ADVERT_ID=$1;")
	if err != nil {
		log.Println(err)
		return
	}
	rows, err := stmt.Query(id)
	if err != nil {
		log.Println(err)
		return
	}
	for rows.Next() {
		var url string
		err = rows.Scan(&url)
		if err != nil {
			log.Println(err)
			return
		}
		urls = append(urls, url)
	}
	return
}

func StoreUserProfileImage(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query()["userID"][0]
	re := regexp.MustCompile(`[\s()]`)
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	wd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}

	err = os.MkdirAll(wd+"/public/images/users_images/"+userID, os.ModeDir)
	if err != nil {
		log.Println(err)
		return
	}
	path := filepath.Join(wd, "public", "images", "users_images", userID, re.ReplaceAllString(fileHeader.Filename, ""))

	newFile, err := os.Create(path)
	if err != nil {
		log.Println(err)
		return
	}
	defer newFile.Close()

	file.Seek(0, 0)
	if _, err := io.Copy(newFile, file); err != nil {
		log.Println(err)
		return
	}

	log.Println("Profile image uploaded successfuly")
}

// StoreNewImage store image uploaded from front end to the server
func StoreNewImage(w http.ResponseWriter, r *http.Request) {
	adID := mux.Vars(r)["id"]
	re := regexp.MustCompile(`[\s()]`)

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	wd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}

	err = os.MkdirAll(wd+"/public/images/adverts/"+adID, os.ModeDir)
	if err != nil {
		log.Println(err)
		return
	}
	path := filepath.Join(wd, "public", "images", "adverts", adID, adID+"_"+re.ReplaceAllString(fileHeader.Filename, ""))

	newFile, err := os.Create(path)
	if err != nil {
		log.Println(err)
		return
	}
	defer newFile.Close()

	file.Seek(0, 0)
	if _, err := io.Copy(newFile, file); err != nil {
		log.Println(err)
		return
	}

	log.Println("image uploaded successfuly")
}

func DeleteAd(w http.ResponseWriter, r *http.Request) {
	CheckUserStatus(w, r)
	adID, err := strconv.Atoi(r.URL.Query()["id"][0])
	if err != nil {
		log.Println(err)
		return
	}
	err = deleteAdByID(adID)
	if err != nil {
		log.Println(err)
		return
	}
}

func deleteAdByID(ID int) (err error) {
	err = deleteAdFromFavoriteTable(ID)
	if err != nil {
		log.Println(err)
	}
	err = deleteAdImagesUrlsFromStorageAndDB(ID)
	if err != nil {
		log.Println(err)
		//return
	}
	stmt, err := Db.Prepare("DELETE FROM ADVERTS WHERE ID=$1")
	if err != nil {
		return err
	}
	res, err := stmt.Exec(ID)
	if err != nil {
		return err
	}
	row, err := res.RowsAffected()
	if row == 1 && err == nil {
		log.Println(row, "annonce supprimé avec succes")
		return
	}
	return errors.New("Can't remove this advert...")
}

// func removeAdFromFavorite(uid) {

// }

func deleteAdFromFavoriteTable(ID int) (err error) {
	_, err = Db.Exec("DELETE FROM FAVORITES WHERE ADVERT_ID=$1", ID)
	if err != nil {
		return
	}
	return
}

func deleteAdImagesUrlsFromStorageAndDB(ID int) (err error) {
	var UID string
	stmt1, err := Db.Prepare("SELECT AD_UID FROM ADVERTS WHERE ID=$1")
	if err != nil {
		return
	}
	err = stmt1.QueryRow(ID).Scan(&UID)
	if err != nil {
		return
	}
	err = deleteImageFromStorage(UID)
	if err != nil {
		return
	}
	stmt2, err := Db.Prepare("DELETE FROM IMAGES_URL WHERE ADVERT_ID=$1")
	if err != nil {
		return
	}
	res, err := stmt2.Exec(ID)
	if err != nil {
		return
	}
	row, err := res.RowsAffected()
	if row > 0 && err == nil {
		log.Println(row, "url supprimé avec succes")

		return err
	}
	return errors.New("Can't remove these images...")
}

func deleteImageFromStorage(UID string) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}
	err = os.RemoveAll(wd + "/public/images/adverts/" + UID)
	if err != nil {
		log.Println("Can't remove this image from local storage...")
		return
	}
	return
}

func UpdateAd(w http.ResponseWriter, r *http.Request) {
	var advert Advert
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(bs, &advert)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = advert.updateAdByID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *Advert) updateAdByID() (err error) {
	stmt, err := Db.Prepare("UPDATE ADVERTS SET LOCATION=$1, OWNER_ID=$2, TITLE=$3, DESCRIPTION=$4, CATEGORY=$5, PRICE=$6, CONTACT=$7, CREATED_AT=$8, AD_UID=$9, ADDRESS=$10 WHERE ID=$11")
	if err != nil {
		return
	}
	res, err := stmt.Exec(a.Location, a.OwnerID, a.Title, a.Description, a.Category, a.Price, a.Contact, a.CreatedAt, a.UID, a.Address, a.ID)
	if err != nil {
		return
	}
	if row, err := res.RowsAffected(); row == 1 && err == nil {
		log.Println("Updated successfully...")
		return err
	}
	log.Println("Can't update advert...")
	return
}
