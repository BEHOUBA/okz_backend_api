package models

import (
	"database/sql"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // importing postgresql driver to deal with database
	"github.com/satori/go.uuid"
)

// Advert struct representing each ad data
type Advert struct {
	OwnerID     int    `json:"userID"`
	UID         string    `json:"Uid"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Price       string    `json:"price"`
	ImgURL      []string  `json:"imgUrls"`
	Contact     string    `json:"contact"`
	Address     string    `json:"address"`
	CreatedAt   time.Time `json:"createdAt"`
	IsFavorite  bool      `json:"isFavorite"`
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
	stmt, err := Db.Prepare("INSERT INTO ADVERTS (LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id;")
	if err != nil {
		log.Println(err)
		return
	}

	imgStmt, err := Db.Prepare("INSERT INTO IMAGES_URL (ADVERT_ID, URL) VALUES ($1, $2)")
	if err != nil {
		log.Println(err)
		return
	}

	err = stmt.QueryRow(a.Location, a.OwnerID, a.Title, a.Description, a.Category, a.Price, a.Contact, time.Now(), a.UID).Scan(&newAdID)
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
	// r.ParseMultipartForm(10000)
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
	advert, err := getAdvertFromDBByUID(adUID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	bs, err := json.Marshal(advert)
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(bs)
}

func getAdvertFromDBByUID(UID string) (advert Advert, err error) {
	stmt, err := Db.Prepare("SELECT ID, LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID FROM ADVERTS WHERE AD_UID=$1")
	if err != nil {
		return
	}
	var id int
	err = stmt.QueryRow(UID).Scan(&id, &advert.Location, &advert.OwnerID, &advert.Title, &advert.Description, &advert.Category, &advert.Price, &advert.Contact, &advert.CreatedAt, &advert.UID)
	if err != nil {
		return
	}
	advert.ImgURL, err = getAdvertImagesURL(id)
	if err != nil {
		log.Println(err, "failed in gettin images's urls")
		return
	}
	return
}

func getAdvertByID(ID int) (advert Advert, err error) {
	stmt, err := Db.Prepare("SELECT ID, LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID FROM ADVERTS WHERE ID=$1")
	if err != nil {
		return
	}
	var id int
	err = stmt.QueryRow(ID).Scan(&id, &advert.Location, &advert.OwnerID, &advert.Title, &advert.Description, &advert.Category, &advert.Price, &advert.Contact, &advert.CreatedAt, &advert.UID)
	if err != nil {
		return
	}
	advert.ImgURL, err = getAdvertImagesURL(id)
	if err != nil {
		log.Println(err, "failed in gettin images's urls")
		return
	}
	return
}

func getAdvertsByUserID(ID int) (ads []Advert, err error) {
	stmt, err := Db.Prepare("SELECT ID, LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID FROM ADVERTS WHERE OWNER_ID=$1")
	if err != nil {
		return
	}
	rows, err := stmt.Query(ID)
	if err != nil {
		return
	}
	for rows.Next() {
		var id int
		var advert Advert
		err = rows.Scan(&id, &advert.Location, &advert.OwnerID, &advert.Title, &advert.Description, &advert.Category, &advert.Price, &advert.Contact, &advert.CreatedAt, &advert.UID)
		if err != nil {
			return
		}
		advert.ImgURL, err = getAdvertImagesURL(id)
		if err != nil {
			log.Println(err, "failed in gettin images's urls")
		}
		ads = append(ads, advert)
	}
	return
}

func getFavoritesByUserID(ID int) (ads []Advert, err error) {
	stmt, err := Db.Prepare("SELECT ADVERT_UID FROM FAVORITES WHERE USER_ID=$1;")
	if err != nil {
		return
	}
	stmt2, err := Db.Prepare("SELECT ID, LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID FROM ADVERTS WHERE AD_UID=$1")
	if err != nil {
		return
	}
	rows, err := stmt.Query(ID)
	if err != nil {
		return
	}
	for rows.Next() {
		var adUID string
		var advert Advert
		var id int
		err = rows.Scan(&adUID)
		if err != nil {
			return
		}
		err = stmt2.QueryRow(adUID).Scan(&id, &advert.Location, &advert.OwnerID, &advert.Title, &advert.Description, &advert.Category, &advert.Price, &advert.Contact, &advert.CreatedAt, &advert.UID)
		if err != nil {
			return
		}
		advert.ImgURL, err = getAdvertImagesURL(id)
		if err != nil {
			log.Println(err, "failed in gettin images's urls")
		}
		ads = append(ads, advert)
	}
	return
}

func getAdvertsFromDB(limit, offset int) (ads []Advert, err error) {
	stmt, err := Db.Prepare("SELECT ID, LOCATION, OWNER_ID, TITLE, DESCRIPTION, CATEGORY, PRICE, CONTACT, CREATED_AT, AD_UID FROM ADVERTS ORDER BY CREATED_AT LIMIT $1 OFFSET $2")
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
		var ID int
		err = rows.Scan(&ID, &ad.Location, &ad.OwnerID, &ad.Title, &ad.Description, &ad.Category, &ad.Price, &ad.Contact, &ad.CreatedAt, &ad.UID)
		if err != nil {
			log.Println(err)
			return
		}
		ad.ImgURL, err = getAdvertImagesURL(ID)
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

// StoreNewImage store image uploaded from front end to the server
func StoreNewImage(w http.ResponseWriter, r *http.Request) {

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
	adID := mux.Vars(r)["id"]
	path := filepath.Join(wd, "public", "images", adID+"_"+fileHeader.Filename)

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

	log.Println("succed")
}
