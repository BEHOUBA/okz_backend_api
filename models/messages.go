package models

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// Message struct for messaging beetween users
type Message struct {
	ID        int
	FromUser  int    `json:"fromUser"`
	ToUser    int    `json:"toUser"`
	AdvertUID string `json:"advertUID"`
	Body      string `json:"body"`
}

func MessageReceiver(w http.ResponseWriter, r *http.Request) {
	var message Message
	bs := make([]byte, r.ContentLength)

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	err = json.Unmarshal(bs, &message)
	if err != nil {
		log.Println(err)
		return
	}
	err = message.saveMessageToDB()
	if err != nil {
		log.Println(err)
		return
	}
}

func (m *Message) saveMessageToDB() (err error) {
	stmt, err := Db.Prepare("INSERT INTO MESSAGES (SENDER_ID, RECEIVER_ID, BODY, ADVERT_UID) VALUES ($1, $2, $3, $4);")
	if err != nil {
		return
	}
	_, err = stmt.Exec(m.FromUser, m.ToUser, m.Body, m.AdvertUID)
	if err != nil {
		return
	}
	log.Println("message store to database successfully !")
	return
}
