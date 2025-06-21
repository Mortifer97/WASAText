package database

import (
	"fmt"
	"time"
)

// UpdateUserPhoto updates the photo for a given user
func (db *appdbimpl) UpdateUserPhoto(userId int64, photoData []byte) error {
	// Prepare the SQL query to update the user's photo
	_, err := db.c.Exec(`
		UPDATE users 
		SET photo = ? 
		WHERE id = ?`, photoData, userId)
	if err != nil {
		return fmt.Errorf("error updating user photo: %w", err)
	}

	return nil
}

// UpdateGroupPhoto updates the photo for a given group
func (db *appdbimpl) UpdateGroupPhoto(conversationId int64, photoData []byte) error {
	// Prepare the SQL query to update the group's photo
	_, err := db.c.Exec(`
		UPDATE conversations 
		SET photo = ? 
		WHERE conversation_id = ?`, photoData, conversationId)
	if err != nil {
		return fmt.Errorf("error updating group photo: %w", err)
	}

	return nil
}

// UpdateGroupName change the name of an exsiting group
func (db *appdbimpl) UpdateGroupName(conversationId int64, newName string) error {
	_, err := db.c.Exec("UPDATE conversations SET name = ? WHERE conversation_id = ?", newName, conversationId)
	if err != nil {
		return fmt.Errorf("failed to update group name: %w", err)
	}
	return nil
}

// UpdateUsername update the name of a specified user
func (db *appdbimpl) UpdateUsername(userId int64, newUsername string) error {
	_, err := db.c.Exec("UPDATE users SET name = ? WHERE id = ?", newUsername, userId)
	return err
}

func (db *appdbimpl) UpdateLastAccess(userId int64, conversationId int64) error {
	if userId <= 0 || conversationId <= 0 {
		return fmt.Errorf("invalid userId (%d) or conversationId (%d)", userId, conversationId)
	}

	query := `
		UPDATE conversation_members
		SET last_access = ?
		WHERE user_id = ? AND conversation_id = ?;
	`

	timestamp := time.Now()
	result, err := db.c.Exec(query, timestamp, userId, conversationId)
	if err != nil {
		return fmt.Errorf("failed to update last_access: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated, possibly invalid userId (%d) or conversationId (%d)", userId, conversationId)
	}

	return nil
}

// UpdateMessageStatus updates messages status
func (db *appdbimpl) UpdateMessagesStatus(conversationId int64, messages []Message) ([]Message, error) {
	// Get the minimum timestamp in conversation_members for the conversation
	var minTimestampStr string
	err := db.c.QueryRow(`
		SELECT CASE
			WHEN COUNT(last_access) < COUNT(*) THEN '1970-01-01 00:00:00'
			ELSE MIN(last_access)
		END
		FROM conversation_members
		WHERE conversation_id = ?`, conversationId).Scan(&minTimestampStr)
	if err != nil {
		return nil, fmt.Errorf("error retrieving minimum last_access timestamp: %w", err)
	}

	// Use ParseTimestamp to transform the string in time.Time
	minTimestamp, err := ParseTimestamp(minTimestampStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing timestamp: %w", err)
	}

	// Update the status of each message
	for i, msg := range messages {
		if msg.Timestamp.Before(minTimestamp) || msg.Timestamp.Equal(minTimestamp) {
			messages[i].Status = "read"
		} else {
			messages[i].Status = "received"
		}
	}

	return messages, nil
}

// UpdateLastMessageId update the filed last_message_id in the conversations table
func (db *appdbimpl) UpdateLastMessageId(conversationId int64, messageId int64) error {
	_, err := db.c.Exec(`
		UPDATE conversations
		SET last_message_id = ?
		WHERE conversation_id = ?`, messageId, conversationId)
	if err != nil {
		return fmt.Errorf("error updating last_message_id: %w", err)
	}
	return nil
}
