package models

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type citiesAndCategories struct {
	Cities     []string `json:"cities"`
	Categories []string `json:"categories"`
}
type categorieWithCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func GetAds(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limitS := query["limit"][0]
	offsetS := query["offset"][0]
	location := strings.Replace(query["location"][0], "_", " ", -1)
	category := strings.Replace(query["category"][0], "_", " ", -1)
	input := strings.Join(query["input"], "")
	sort := strings.Join(query["sort"], "")
	log.Println(location, category)

	limit, err := strconv.Atoi(limitS)
	if err != nil {
		log.Println(err)
	}
	offset, err := strconv.Atoi(offsetS)
	if err != nil {
		log.Println(err)
	}
	result, err := searchAdverts(location, category, input, sort, limit, offset)
	if err != nil {
		log.Println(err)
		return
	}
	bs, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(bs)
}

func searchAdverts(location, category, input, sort string, limit, offset int) (adverts []Advert, err error) {
	var stmt *sql.Stmt
	if location == "" && category == "" && input == "" && sort == "" {
		stmt, err = Db.Prepare("SELECT * FROM ADVERTS ORDER BY CREATED_AT LIMIT $1 OFFSET $2")
		if err != nil {
			log.Println(err)
			return
		}
		return getAdvertsFromDB(limit, offset)
	} else {
		switch sort {
		case "prix_croissant":
			stmt, err = Db.Prepare("SELECT * FROM ADVERTS WHERE LOWER(LOCATION) LIKE LOWER('%' || $1 ||'%') AND LOWER(CATEGORY) LIKE LOWER('%' || $2 ||'%') AND LOWER(DESCRIPTION) LIKE ('%' || $3 ||'%') AND LOWER(TITLE) LIKE ('%' || $4 ||'%') ORDER BY PRICE ASC LIMIT $5 OFFSET $6;")
		case "prix_decroissant":
			stmt, err = Db.Prepare("SELECT * FROM ADVERTS WHERE LOWER(LOCATION) LIKE LOWER('%' || $1 ||'%') AND LOWER(CATEGORY) LIKE LOWER('%' || $2 ||'%') AND LOWER(DESCRIPTION) LIKE ('%' || $3 ||'%') AND LOWER(TITLE) LIKE ('%' || $4 ||'%') ORDER BY PRICE DESC LIMIT $5 OFFSET $6;")
		case "nouveautés":
			stmt, err = Db.Prepare("SELECT * FROM ADVERTS WHERE LOWER(LOCATION) LIKE LOWER('%' || $1 ||'%') AND LOWER(CATEGORY) LIKE LOWER('%' || $2 ||'%') AND LOWER(DESCRIPTION) LIKE ('%' || $3 ||'%') AND LOWER(TITLE) LIKE ('%' || $4 ||'%') ORDER BY CREATED_AT DESC LIMIT $5 OFFSET $6;")
		default:
			stmt, err = Db.Prepare("SELECT * FROM ADVERTS WHERE LOWER(LOCATION) LIKE LOWER('%' || $1 ||'%') AND LOWER(CATEGORY) LIKE LOWER('%' || $2 ||'%') AND LOWER(DESCRIPTION) LIKE ('%' || $3 ||'%') AND LOWER(TITLE) LIKE ('%' || $4 ||'%') LIMIT $5 OFFSET $6;")
		}
	}
	if err != nil {
		log.Println(err)
		return
	}
	rows, err := stmt.Query(location, category, input, input, limit, offset)
	if err != nil {
		log.Println(err)
		return
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
		adverts = append(adverts, ad)
	}
	return
}

func GetCitiesAndCategoriesInDB(w http.ResponseWriter, r *http.Request) {
	var cities, categories []string
	var err error
	cities, err = getCitiesInDB()
	if err != nil {
		return
	}
	categories, err = getCategoriesInDB()
	if err != nil {
		return
	}
	response := citiesAndCategories{cities, categories}
	log.Println(response)
	bs, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(bs)
}

func getCitiesInDB() (cities []string, err error) {
	stmt, err := Db.Prepare("SELECT DISTINCT LOCATION FROM ADVERTS")
	if err != nil {
		log.Println(err)
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Println(err)
		return
	}
	for rows.Next() {
		var city string
		rows.Scan(&city)
		cities = append(cities, city)
	}
	return
}

func getCategoriesInDB() (categories []string, err error) {
	stmt, err := Db.Prepare("SELECT DISTINCT CATEGORY FROM ADVERTS")
	if err != nil {
		log.Println(err)
		return
	}

	rows, err := stmt.Query()
	if err != nil {
		log.Println(err)
		return
	}

	for rows.Next() {
		var cat string
		rows.Scan(&cat)
		categories = append(categories, cat)
	}
	return
}

func AdvertByCategoriesAndCount(w http.ResponseWriter, r *http.Request) {
	var categoriesWithCount []categorieWithCount
	categories, err := getCategoriesInDB()
	for _, val := range categories {
		var count int
		row := Db.QueryRow("SELECT COUNT(CATEGORY) FROM ADVERTS WHERE CATEGORY=$1;", val)
		row.Scan(&count)
		categoriesWithCount = append(categoriesWithCount, categorieWithCount{val, count})
	}
	bs, err := json.Marshal(categoriesWithCount)
	if err != nil {
		log.Println(err)
		return
	}
	w.Write(bs)
}
