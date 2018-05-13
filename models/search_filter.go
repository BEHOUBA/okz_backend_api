package models

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	location := strings.Join(query["location"], "")
	category := strings.Join(query["category"], "")
	input := strings.Join(query["input"], "")
	log.Println(location, category, input)

	result, err := searchAdverts(location, category, input)
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

func searchAdverts(location, category, input string) (adverts []Advert, err error) {
	stmt, err := Db.Prepare("SELECT * FROM ADVERTS WHERE LOWER(LOCATION) LIKE LOWER('%' || $1 ||'%') OR LOWER(CATEGORY) LIKE LOWER('%' || $2 ||'%') OR LOWER(DESCRIPTION) LIKE ('%' || $3 ||'%') OR LOWER(TITLE) LIKE ('%' || $4 ||'%');")
	if err != nil {
		log.Println(err)
		return
	}
	rows, err := stmt.Query(location, category, input, input)
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
