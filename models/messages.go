package models

// Message struct for messaging beetween users
type Message struct {
	ID         int
	SenderID   int
	ReceiverID int
	Body       string
}
