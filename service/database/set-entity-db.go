package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Funzione per aggiungere un commento a un messaggio.
// Si collega agli handler commentMessage e uncommentMessage.
func (db *appdbimpl) AddCommentToMessage(messageId int64, senderId int64, content string) (Comment, error) {
	sender, err := db.GetUserById(senderId)
	if err != nil {
		return Comment{}, fmt.Errorf("errore recupero sender: %w", err)
	}

	result, err := db.c.Exec(
		"INSERT INTO comments (message_id, sender_id, content) VALUES (?, ?, ?)",
		messageId, senderId, content,
	)
	if err != nil {
		return Comment{}, fmt.Errorf("errore inserimento commento: %w", err)
	}

	commentId, err := result.LastInsertId()
	if err != nil {
		return Comment{}, fmt.Errorf("errore recupero id commento: %w", err)
	}

	return Comment{
		CommentId: commentId,
		Content:   content,
		Sender:    sender,
	}, nil
}

func (db *appdbimpl) AddUserToGroup(conversationId int64, userId int64) error {
	_, err := db.c.Exec(
		`INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`,
		conversationId, userId,
	)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}
	return nil
}

// CreateConversation creates a new conversation between two users
func (db *appdbimpl) CreateConversation(user1Id, user2Id int64, conversationType string) (Conversation, error) {
	// Start a transaction to ensure atomicity
	tx, err := db.c.Begin()
	if err != nil {
		return Conversation{}, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { // Ensure rollback in case of failure
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			fmt.Errorf("error rolling back transaction: %w", err)
		}
	}()

	// Determine the name of the conversation based on type
	var conversationName string
	if conversationType == "direct" {
		// Get the target user's name for direct conversation
		targetUser, err := db.GetUserById(user2Id)
		if err != nil {
			return Conversation{}, fmt.Errorf("error fetching target user: %w", err)
		}
		conversationName = targetUser.Name
	} else {
		conversationName = "New Conversation"
	}

	// Create the new conversation
	result, err := tx.Exec(`
		INSERT INTO conversations (name, last_message_id, type)
		VALUES (?, ?, ?)
	`, conversationName, nil, conversationType)
	if err != nil {
		return Conversation{}, fmt.Errorf("error inserting conversation: %w", err)
	}

	// Get the conversation ID
	conversationId, err := result.LastInsertId()
	if err != nil {
		return Conversation{}, fmt.Errorf("error retrieving conversation ID: %w", err)
	}

	// Add the two users to the conversation
	_, err = tx.Exec(`
		INSERT INTO conversation_members (conversation_id, user_id)
		VALUES (?, ?), (?, ?)
	`, conversationId, user1Id, conversationId, user2Id)
	if err != nil {
		return Conversation{}, fmt.Errorf("error adding users to conversation: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return Conversation{}, fmt.Errorf("error committing transaction: %w", err)
	}

	// Fetch the conversation
	conversation, err := db.GetConversationById(conversationId)
	if err != nil {
		return Conversation{}, fmt.Errorf("error fetching conversation details: %w", err)
	}

	return conversation, nil
}

// Si collegano agli handler addToGroup, leaveGroup, setGroupName, getGroupMembers.
// Esempio: AddUserToGroup viene chiamata da addToGroup in api/put-user-to-group.go.

func (db *appdbimpl) CreateGroup(conversation Conversation) error {
	_, err := db.c.Exec(
		`INSERT INTO conversations (name, photo_url) VALUES (?, ?)`,
		conversation.Name, conversation.Photo,
	)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}
	return nil
}

// CreateUser create a new user in the database
func (db *appdbimpl) CreateUser(name string) (User, error) {
	result, err := db.c.Exec("INSERT INTO users (name) VALUES (?)", name)
	if err != nil {
		return User{}, fmt.Errorf("error inserting user: %w", err)
	}

	userId, err := result.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("error retrieving last insert id: %w", err)
	}

	return User{
		UserId: userId,
	}, nil
}

// ReplyMessage adds a reply to an existing message with either text or a photo.
func (db *appdbimpl) ReplyMessage(conversationId int64, senderId int64, replyMessageId int64, text string, status string, messageType string, photo []byte) (Message, error) {
	sender, err := db.GetUserById(senderId)
	if err != nil {
		return Message{}, fmt.Errorf("error fetching sender details: %w", err)
	}

	timestamp := time.Now()
	var result sql.Result

	if messageType == "photo" {
		// Handle photo reply message
		result, err = db.c.Exec(`
			INSERT INTO messages (timestamp, text, conversation_id, sender_id, status, type, reply_to_message_id, photo) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, timestamp, "", conversationId, senderId, status, "reply", replyMessageId, photo)
	} else {
		// Handle text reply message
		result, err = db.c.Exec(`
			INSERT INTO messages (timestamp, text, conversation_id, sender_id, status, type, reply_to_message_id) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`, timestamp, text, conversationId, senderId, status, "reply", replyMessageId)
	}

	if err != nil {
		return Message{}, fmt.Errorf("error inserting reply message: %w", err)
	}

	messageId, err := result.LastInsertId()
	if err != nil {
		return Message{}, fmt.Errorf("error retrieving last insert id: %w", err)
	}

	// Update the conversation's last_message_id
	if err := db.UpdateLastMessageId(conversationId, messageId); err != nil {
		return Message{}, fmt.Errorf("error updating last_message_id: %w", err)
	}

	var msg Message
	if messageType == "photo" {
		msg.Photo = photo
	} else {
		msg.Text = text
	}

	return Message{
		MessageId:        messageId,
		Timestamp:        timestamp,
		Text:             msg.Text,
		Sender:           sender,
		Status:           status,
		Type:             "reply",
		ReplyToMessageId: &replyMessageId,
		Photo:            msg.Photo,
	}, nil
}

// ForwardMessage forward a message into a conversation.
func (db *appdbimpl) ForwardMessage(userId int64, originalMessage Message, targetConversationId int64) (Message, error) {
	// Retrieve sender details
	sender, err := db.GetUserById(userId)
	if err != nil {
		return Message{}, fmt.Errorf("error retrieving sender: %w", err)
	}
	timestamp := time.Now()
	// Prepare the query to insert the forwarded message
	var result sql.Result
	if originalMessage.Photo != nil {
		// If the original message has a photo, include it in the insert
		result, err = db.c.Exec(`
			INSERT INTO messages (timestamp, text, conversation_id, sender_id, status, type, photo)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			timestamp, originalMessage.Text, targetConversationId, userId, "received", "forward", originalMessage.Photo)
	} else {
		// If no photo, proceed with only the text content
		result, err = db.c.Exec(`
			INSERT INTO messages (timestamp, text, conversation_id, sender_id, status, type)
			VALUES (?, ?, ?, ?, ?, ?)`,
			timestamp, originalMessage.Text, targetConversationId, userId, "received", "forward")
	}
	if err != nil {
		return Message{}, fmt.Errorf("error forwarding message: %w", err)
	}

	newMessageId, err := result.LastInsertId()
	if err != nil {
		return Message{}, fmt.Errorf("error retrieving last insert id: %w", err)
	}

	// Update the last_message_id of the target conversation
	if err := db.UpdateLastMessageId(targetConversationId, newMessageId); err != nil {
		return Message{}, fmt.Errorf("error updating last_message_id for target conversation: %w", err)
	}

	// Create the new forwarded message
	forwardedMessage := Message{
		MessageId: newMessageId,
		Timestamp: timestamp,
		Text:      originalMessage.Text,
		Sender:    sender,
		Status:    "received",
		Type:      "forward",
	}

	// If the original message contained a photo, include it in the forwarded message
	if originalMessage.Photo != nil {
		forwardedMessage.Photo = originalMessage.Photo
	}

	return forwardedMessage, nil
}

// AddMessage adds a new message to the database.
func (db *appdbimpl) AddMessage(conversationId int64, senderId int64, text string, status string, messageType string, photo []byte) (Message, error) {
	sender, err := db.GetUserById(senderId)
	if err != nil {
		return Message{}, fmt.Errorf("error fetching sender details: %w", err)
	}

	timestamp := time.Now()
	var result sql.Result
	if messageType == "photo" {
		// Handle photo message
		result, err = db.c.Exec(`
			INSERT INTO messages (timestamp, text, conversation_id, sender_id, status, type, photo) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`, timestamp, text, conversationId, senderId, status, "standard", photo)
	} else {
		// Handle text message
		result, err = db.c.Exec(`
			INSERT INTO messages (timestamp, text, conversation_id, sender_id, status, type) 
			VALUES (?, ?, ?, ?, ?, ?)`, timestamp, text, conversationId, senderId, status, "standard")
	}

	if err != nil {
		return Message{}, fmt.Errorf("error inserting message: %w", err)
	}

	messageId, err := result.LastInsertId()
	if err != nil {
		return Message{}, fmt.Errorf("error retrieving last insert id: %w", err)
	}

	// Update the conversation's last_message_id
	if err := db.UpdateLastMessageId(conversationId, messageId); err != nil {
		return Message{}, fmt.Errorf("error updating last_message_id: %w", err)
	}

	var msg Message
	if messageType == "photo" {
		msg.Photo = photo
	} else {
		msg.Text = text
	}

	return Message{
		MessageId: messageId,
		Timestamp: timestamp,
		Text:      text,
		Sender:    sender,
		Status:    status,
		Type:      "standard",
		Photo:     msg.Photo,
	}, nil
}

// ParseTimestamp converts a datetime string to time.Time
func ParseTimestamp(datetimeStr string) (time.Time, error) {
	if datetimeStr == "" {
		// Return a zero value of time.Time if the string is empty
		return time.Time{}, nil
	}

	// Handle the specific case for the "epoch" timestamp without timezone
	if datetimeStr == "1970-01-01 00:00:00" {
		// Return the zero value of time.Time (1970-01-01 00:00:00 UTC)
		return time.Time{}, nil
	}

	// Timestamp format that includes the timezone
	const layout = "2006-01-02 15:04:05.999999999-07:00"
	parsedTime, err := time.Parse(layout, datetimeStr)
	if err != nil {
		// Return an error if parsing fails
		return time.Time{}, fmt.Errorf("error parsing timestamp: %w", err)
	}
	return parsedTime, nil
}
