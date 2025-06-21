package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// Funzioni per la gestione delle conversazioni.
// GetConversationsByUser recupera tutte le conversazioni di un utente.
// Si collega agli handler getConversations e addConversation.
func (db *appdbimpl) GetConversationsByUser(userId int64, sortOrder string) ([]Conversation, error) {
	query := `
		SELECT 
			c.conversation_id, 
			c.name, 
			c.photo, 
			COALESCE(m.message_id, 0) AS message_id,
    		COALESCE(m.timestamp, NULL) AS timestamp,
    		COALESCE(m.text, '') AS content,
			c.type
		FROM 
			conversations c
		INNER JOIN 
			conversation_members cm 
		ON 
			c.conversation_id = cm.conversation_id
		LEFT JOIN 
			messages m 
		ON 
			c.last_message_id = m.message_id
			AND c.conversation_id = m.conversation_id
		WHERE 
			cm.user_id = ?
		ORDER BY 
			m.timestamp ` + sortOrder
	rows, err := db.c.Query(query, userId)
	if err != nil {
		return nil, fmt.Errorf("errore recupero conversazioni: %w", err)
	}
	defer rows.Close()
	var conversations []Conversation
	for rows.Next() {
		var conversation Conversation
		var lastMessage LastMessage
		var timestampStr *string
		var photo []byte
		err := rows.Scan(
			&conversation.ConversationId,
			&conversation.Name,
			&photo,
			&lastMessage.MessageID,
			&timestampStr,
			&lastMessage.Preview,
			&conversation.Type,
		)
		if err != nil {
			return nil, fmt.Errorf("errore scan conversazione: %w", err)
		}
		if len(photo) == 0 {
			conversation.Photo = nil
		} else {
			conversation.Photo = photo
		}
		if timestampStr != nil {
			lastMessage.Timestamp, err = ParseTimestamp(*timestampStr)
			if err != nil {
				return nil, fmt.Errorf("errore parsing timestamp: %w", err)
			}
		}
		if lastMessage.MessageID == 0 {
			conversation.LastMessage = nil
		} else {
			conversation.LastMessage = &lastMessage
		}
		// Personalizzazione per conversazioni dirette: mostra nome/foto dell'altro utente
		if conversation.Type == "direct" {
			otherUser, err := db.GetOtherUserInConversation(conversation.ConversationId, userId)
			if err != nil {
				return nil, fmt.Errorf("errore recupero altro utente nella conversazione diretta: %w", err)
			}
			conversation.Name = otherUser.Name
			if otherUser.Photo != nil && len(otherUser.Photo) > 0 {
				conversation.Photo = otherUser.Photo
			}
		}
		conversations = append(conversations, conversation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("errore iterazione conversazioni: %w", err)
	}
	return conversations, nil
}

// GetConversationById retrieves the conversation details by its ID
func (db *appdbimpl) GetConversationById(conversationId int64) (Conversation, error) {
	// Query to retrieve the conversation details
	query := `
		SELECT 
			c.conversation_id, 
			c.name, 
			c.photo, 
			COALESCE(m.message_id, 0) AS message_id, 
			COALESCE(m.timestamp, NULL) AS timestamp, 
			COALESCE(m.text, '') AS content,
			c.type
		FROM 
			conversations c
		LEFT JOIN 
			messages m 
		ON 
			c.last_message_id = m.message_id
		WHERE 
			c.conversation_id = ?`

	var conversation Conversation
	var lastMessage LastMessage
	var timestampStr *string
	var photo sql.NullByte

	// Execute the query
	row := db.c.QueryRow(query, conversationId)

	// Scan the result into the Conversation structure
	err := row.Scan(
		&conversation.ConversationId,
		&conversation.Name,
		&photo,
		&lastMessage.MessageID,
		&timestampStr,
		&lastMessage.Preview,
		&conversation.Type,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Conversation{}, fmt.Errorf("conversation not found")
		}
		return Conversation{}, fmt.Errorf("error scanning conversation: %w", err)
	}

	// If photo is not NULL, set it to the conversation's Photo field
	if photo.Valid {
		conversation.Photo = []byte{photo.Byte}
	}

	// Parsing timestamp
	if timestampStr != nil {
		lastMessage.Timestamp, err = ParseTimestamp(*timestampStr)
		if err != nil {
			return Conversation{}, fmt.Errorf("error parsing timestamp: %w", err)
		}
	}

	// Ad the message only if exists
	if lastMessage.MessageID == 0 {
		conversation.LastMessage = nil
	} else {
		conversation.LastMessage = &lastMessage
	}

	return conversation, nil
}

// GetOtherUserInConversation retrieves the other user in a direct conversation
func (db *appdbimpl) GetOtherUserInConversation(conversationId, userId int64) (*User, error) {
	var otherUser User
	var photo []byte

	// Query to find the other user in the conversation
	query := `
		SELECT u.id, u.name, u.photo
		FROM conversation_members cm
		INNER JOIN users u ON cm.user_id = u.id
		WHERE cm.conversation_id = ? AND cm.user_id != ?
	`

	err := db.c.QueryRow(query, conversationId, userId).Scan(
		&otherUser.UserId,
		&otherUser.Name,
		&photo,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("no other user found in conversation %d for user %d", conversationId, userId)
	}
	if err != nil {
		return nil, fmt.Errorf("error retrieving other user: %w", err)
	}

	// If photo is valid, assign the photo as a byte slice
	if len(photo) == 0 {
		otherUser.Photo = nil
	} else {
		otherUser.Photo = photo
	}

	return &otherUser, nil
}

func (db *appdbimpl) GetGroupById(conversationId int64) (*Conversation, error) {
	row := db.c.QueryRow(
		`SELECT conversation_id, name, photo FROM conversations WHERE conversation_id = ?`,
		conversationId,
	)
	var conversation Conversation
	if err := row.Scan(&conversation.ConversationId, &conversation.Name, &conversation.Photo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("group not found: %w", err)
		}
		return nil, fmt.Errorf("error retrieving group: %w", err)
	}
	return &conversation, nil
}

// GetCommentById retrieves a comment by its ID.
func (db *appdbimpl) GetCommentById(commentId int64) (Comment, error) {
	var comment Comment
	var senderId int64

	query := `
		SELECT comment_id, sender_id, content
		FROM comments
		WHERE comment_id = ?
	`
	err := db.c.QueryRow(query, commentId).Scan(&comment.CommentId, &senderId, &comment.Content)
	if err != nil {
		return Comment{}, fmt.Errorf("error retrieving comment: %w", err)
	}

	// Fetch user details from the database
	sender, err := db.GetUserById(senderId)
	if err != nil {
		return Comment{}, fmt.Errorf("error fetching sender details: %w", err)
	}
	comment.Sender = sender
	return comment, nil
}

func (db *appdbimpl) GetGroupMembers(groupId int64) ([]string, error) {
	// Query to get group members
	rows, err := db.c.Query(`
		SELECT u.name
		FROM users u
		INNER JOIN conversation_members cm ON u.id = cm.user_id
		WHERE cm.conversation_id = ?
	`, groupId)
	if err != nil {
		return nil, fmt.Errorf("failed to query group members: %w", err)
	}
	defer rows.Close()

	// Read the query results
	var members []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, fmt.Errorf("failed to scan group member: %w", err)
		}
		members = append(members, username)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over group members: %w", err)
	}

	return members, nil
}

// IsUserMemberOfGroup checks if a user is a member of a group
func (db *appdbimpl) IsUserMemberOfGroup(userId int64, conversationId int64) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM conversation_members WHERE conversation_id = ? AND user_id = ?)"
	err := db.c.QueryRow(query, conversationId, userId).Scan(&exists)
	return exists, err
}

// GetUserById retrieves a user from the database by their ID
func (db *appdbimpl) GetUserById(userId int64) (User, error) {
	var user User
	var photo []byte

	// Query to retrieve the user details
	err := db.c.QueryRow(`
		SELECT id, name, photo
		FROM users
		WHERE id = ?`, userId).Scan(&user.UserId, &user.Name, &photo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, fmt.Errorf("user not found: %w", err)
		}
		return User{}, fmt.Errorf("error retrieving user: %w", err)
	}

	// If photo is valid, assign the photo as a byte slice
	if len(photo) == 0 {
		user.Photo = nil
	} else {
		user.Photo = photo
	}

	return user, nil
}

// IsUserInConversation checks if a user is part of a conversation.
func (db *appdbimpl) IsUserInConversation(userId int64, conversationId int64) (bool, error) {
	var exists bool
	err := db.c.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM conversation_members 
			WHERE user_id = ? AND conversation_id = ?
		)`, userId, conversationId).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if user is in conversation: %w", err)
	}
	return exists, nil
}

// SearchUsersByUsername searches for users whose username contains the search string
func (db *appdbimpl) SearchUsersByUsername(username string) ([]User, error) {
	var query string
	var args []interface{}

	if username == "" {
		// No username provided, return all users
		query = `SELECT id, name, photo FROM users`
	} else {
		// Search for users whose name contains the search string
		query = `SELECT id, name, photo FROM users WHERE name LIKE ? LIMIT 10`
		args = append(args, "%"+username+"%")
	}

	rows, err := db.c.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error searching users by username: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var photo []byte

		if err := rows.Scan(&user.UserId, &user.Name, &photo); err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}

		// If photo is valid, assign it as a byte slice
		if len(photo) > 0 {
			user.Photo = photo
		} else {
			user.Photo = nil
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, nil
}

// GetUserByName search a user by name
func (db *appdbimpl) GetUserByName(name string) (*User, error) {
	var user User
	var photo []byte

	// Query to retrieve the user details
	err := db.c.QueryRow("SELECT id, name, photo FROM users WHERE name = ?", name).Scan(&user.UserId, &user.Name, &photo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, err
	}

	// If photo is valid, assign the photo as a byte slice
	if len(photo) == 0 {
		user.Photo = nil
	} else {
		user.Photo = photo
	}

	return &user, nil
}

// GetMessageById retrieves a message by its ID.
func (db *appdbimpl) GetMessageById(messageId int64, conversationId int64) (Message, error) {
	var msg Message
	var senderId int64
	var replyToMessageId sql.NullInt64
	var photo []byte
	err := db.c.QueryRow(`
		SELECT message_id, timestamp, text, sender_id, status, type, reply_to_message_id, photo
		FROM messages
		WHERE message_id = ? AND conversation_id = ?`, messageId, conversationId).Scan(
		&msg.MessageId, &msg.Timestamp, &msg.Text, &senderId, &msg.Status, &msg.Type, &replyToMessageId, &photo)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Message{}, fmt.Errorf("message not found: %w", err)
		}
		return Message{}, fmt.Errorf("error retrieving message: %w", err)
	}

	// Retrieve sender details
	sender, err := db.GetUserById(senderId)
	if err != nil {
		return Message{}, fmt.Errorf("error retrieving sender: %w", err)
	}
	msg.Sender = sender

	// Handle nullable ReplyToMessageId
	if replyToMessageId.Valid {
		msg.ReplyToMessageId = &replyToMessageId.Int64
	}

	// Store the photo as a byte slice if it exists
	if len(photo) > 0 {
		msg.Photo = photo
	}

	return msg, nil
}

// GetCommentsByMessage retrieves the comments for a specific message.
func (db *appdbimpl) GetCommentsByMessage(messageId int64) ([]Comment, error) {
	rows, err := db.c.Query(`
		SELECT comment_id, sender_id, content
		FROM comments
		WHERE message_id = ?`, messageId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var senderId int64
		if err := rows.Scan(&comment.CommentId, &senderId, &comment.Content); err != nil {
			return nil, fmt.Errorf("error scanning comment: %w", err)
		}

		// Retrieve the sender's user data
		sender, err := db.GetUserById(senderId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("sender not found for comment: %w", err)
			}
			return nil, fmt.Errorf("error retrieving sender for comment: %w", err)
		}
		comment.Sender = sender
		comments = append(comments, comment)

	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return comments, nil
}

// GetMessagesByConversation retrieves the messages for a specific conversation, sorted by timestamp.
func (db *appdbimpl) GetMessagesByConversation(userId int64, conversationId int64, sortOrder string) ([]Message, error) {
	// Prepare the query to retrieve messages from the database, ordered by timestamp.
	var orderBy string
	if sortOrder == "asc" {
		orderBy = "ASC"
	} else {
		orderBy = "DESC"
	}

	rows, err := db.c.Query(`
		SELECT m.message_id, m.timestamp, m.text, m.sender_id, m.status, m.type, m.reply_to_message_id, m.photo
		FROM messages m
		JOIN conversations c ON c.conversation_id = m.conversation_id
		JOIN conversation_members cm ON cm.conversation_id = c.conversation_id
		WHERE cm.user_id = ? AND c.conversation_id = ?
		ORDER BY m.timestamp `+orderBy, userId, conversationId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving messages: %w", err)
	}
	defer rows.Close()

	// Update the last access timestamp for the user in the conversation
	if err := db.UpdateLastAccess(userId, conversationId); err != nil {
		return nil, fmt.Errorf("failed to update last access: %w", err)
	}

	var messages []Message
	for rows.Next() {
		var msg Message
		var senderId int64
		var replyToMessageId sql.NullInt64
		var photo sql.RawBytes
		if err := rows.Scan(&msg.MessageId, &msg.Timestamp, &msg.Text, &senderId, &msg.Status, &msg.Type, &replyToMessageId, &photo); err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}

		// Retrieve the sender's user data
		sender, err := db.GetUserById(senderId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("sender not found for message %d: %w", msg.MessageId, err)
			}
			return nil, fmt.Errorf("error retrieving sender for message %d: %w", msg.MessageId, err)
		}
		msg.Sender = sender

		// Retrieve the comments for the message (if any)
		msg.Comments, err = db.GetCommentsByMessage(msg.MessageId)
		if err != nil {
			return nil, fmt.Errorf("error retrieving comments for message %d: %w", msg.MessageId, err)
		}

		// Handle nullable ReplyToMessageId
		if replyToMessageId.Valid {
			msg.ReplyToMessageId = &replyToMessageId.Int64
		}

		// Store the photo as a byte slice if it exists
		if photo != nil {
			msg.Photo = photo
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Update message status
	messages, err = db.UpdateMessagesStatus(conversationId, messages)
	if err != nil {
		return nil, fmt.Errorf("error updating message statuses: %w", err)
	}

	return messages, nil
}
