package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// DeleteCommentById deletes a comment by its ID.
func (db *appdbimpl) DeleteCommentById(commentId int64) error {
	query := `
		DELETE FROM comments
		WHERE comment_id = ?
	`
	_, err := db.c.Exec(query, commentId)
	if err != nil {
		return fmt.Errorf("error deleting comment: %w", err)
	}
	return nil
}

// RemoveUserFromGroup removes a user from a group
func (db *appdbimpl) RemoveUserFromGroup(conversationId int64, userId int64) error {
	_, err := db.c.Exec("DELETE FROM conversation_members WHERE conversation_id = ? AND user_id = ?", conversationId, userId)
	return err
}

// Function to delete the comments associated with a message
func (db *appdbimpl) deleteCommentsByMessageId(messageId int64) error {
	query := `
		DELETE FROM comments
		WHERE message_id = ?`
	_, err := db.c.Exec(query, messageId)
	if err != nil {
		return fmt.Errorf("error deleting comments: %w", err)
	}
	return nil
}

// Function to delete the message
func (db *appdbimpl) deleteMessage(messageId int64) error {
	query := `
		DELETE FROM messages
		WHERE message_id = ?`
	_, err := db.c.Exec(query, messageId)
	if err != nil {
		return fmt.Errorf("error deleting message: %w", err)
	}
	return nil
}

// DeleteMessageById deletes a message by its ID
func (db *appdbimpl) DeleteMessageById(messageId int64) error {
	// Delete comments associated with the message first
	if err := db.deleteCommentsByMessageId(messageId); err != nil {
		return fmt.Errorf("error deleting comments: %w", err)
	}

	// Check if the deleted message is the last message in any conversation
	var conversationId int64
	err := db.c.QueryRow(`
		SELECT conversation_id
		FROM conversations
		WHERE last_message_id = ?`, messageId).Scan(&conversationId)

	// If no conversation is found with this last_message_id, just delete the message
	if errors.Is(err, sql.ErrNoRows) {
		// Message is not the last message in any conversation
		return db.deleteMessage(messageId)
	} else if err != nil {
		// Error querying for the conversation
		return fmt.Errorf("error checking if message is last message: %w", err)
	}

	// The message is the last message, so we need to update the conversation's last_message_id
	// Delete the message first
	if err := db.deleteMessage(messageId); err != nil {
		return err
	}

	// Find the new last message by timestamp (most recent)
	var newLastMessageId int64
	err = db.c.QueryRow(`
		SELECT message_id
		FROM messages
		WHERE conversation_id = ?
		ORDER BY timestamp DESC
		LIMIT 1`, conversationId).Scan(&newLastMessageId)

	if err != nil {
		// If no other messages exist set last_message_id to NULL
		if errors.Is(err, sql.ErrNoRows) {
			return db.UpdateLastMessageId(conversationId, 0) // Set to NULL or 0
		}
		return fmt.Errorf("error finding new last message: %w", err)
	}

	// Update the conversation with the new last_message_id
	if err := db.UpdateLastMessageId(conversationId, newLastMessageId); err != nil {
		return fmt.Errorf("error updating last_message_id after deletion: %w", err)
	}

	return nil
}
