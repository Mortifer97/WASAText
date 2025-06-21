package api

import (
	"net/http"
	"strconv"

	"github.com/Mortifer97/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// removeComment handles the API request
func (rt *_router) removeComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Extract userId, conversationId, messageId and commentId from path parameters
	userIdStr := ps.ByName("userId")
	userId, errUsr := strconv.ParseInt(userIdStr, 10, 64)
	conversationIdStr := ps.ByName("conversationId")
	messageIdStr := ps.ByName("messageId")
	commentIdStr := ps.ByName("commentId")
	conversationId, err := strconv.ParseInt(conversationIdStr, 10, 64)
	messageId, errMsg := strconv.ParseInt(messageIdStr, 10, 64)
	commentId, errComment := strconv.ParseInt(commentIdStr, 10, 64)

	if err != nil || errUsr != nil || errMsg != nil || errComment != nil || userId <= 0 || conversationId <= 0 || messageId <= 0 || commentId <= 0 {
		ctx.Logger.WithError(err).Error("invalid userId, conversationId, message, or commentId")
		http.Error(w, "Invalid user, conversation, message, or comment id", http.StatusBadRequest)
		return
	}

	// Check if the user exists
	_, err = rt.db.GetUserById(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("user not found")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the user is part of the conversation
	isMember, err := rt.db.IsUserInConversation(userId, conversationId)
	if err != nil {
		ctx.Logger.WithError(err).Error("error checking if user is in conversation")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		ctx.Logger.Error("user is not a member of the conversation")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Verify if the comment exists
	comment, err := rt.db.GetCommentById(commentId)
	if err != nil {
		ctx.Logger.WithError(err).Error("comment not found")
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Verify ownership or permissions (if needed)
	if comment.Sender.UserId != userId {
		ctx.Logger.Error("user unauthorized to delete comment")
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Delete the comment
	err = rt.db.DeleteCommentById(commentId)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to delete comment")
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	// Respond with no content (204)
	w.WriteHeader(http.StatusNoContent)
}

// leaveGroup handles the API request
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Extract userId and groupId from path parameters
	userIdStr := ps.ByName("userId")
	userId, errUsr := strconv.ParseInt(userIdStr, 10, 64)
	groupIdStr := ps.ByName("groupId")
	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil || errUsr != nil || userId <= 0 || groupId <= 0 {
		ctx.Logger.WithError(err).Error("invalid userId or groupId")
		http.Error(w, "Invalid user or group id", http.StatusBadRequest)
		return
	}

	// Check if the user exists
	_, err = rt.db.GetUserById(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("user not found")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the group exists
	_, err = rt.db.GetGroupById(int64(groupId))
	if err != nil {
		ctx.Logger.WithError(err).Error("group not found")
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Check if the user is a member of the group
	isMember, err := rt.db.IsUserMemberOfGroup(userId, int64(groupId))
	if err != nil {
		ctx.Logger.WithError(err).Error("database error checking group membership")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		ctx.Logger.Error("user is not a member of the group")
		http.Error(w, "User is not a member of the group", http.StatusNotFound)
		return
	}

	// Remove the user from the group
	err = rt.db.RemoveUserFromGroup(int64(groupId), userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to remove user from group")
		http.Error(w, "Failed to remove user from group", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusNoContent)
}

// deleteMessage handles the API request.
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Extract userId, conversationId, and messageId from path parameters
	userIdStr := ps.ByName("userId")
	userId, errUsr := strconv.ParseInt(userIdStr, 10, 64)
	conversationIdStr := ps.ByName("conversationId")
	messageIdStr := ps.ByName("messageId")
	conversationId, err := strconv.ParseInt(conversationIdStr, 10, 64)
	messageId, errMsg := strconv.ParseInt(messageIdStr, 10, 64)

	if err != nil || errUsr != nil || errMsg != nil || userId <= 0 || conversationId <= 0 || messageId <= 0 {
		ctx.Logger.WithError(err).Error("invalid userId, conversationId or messageId")
		http.Error(w, "Invalid user, conversation or message id", http.StatusBadRequest)
		return
	}

	// Check if the user exists
	_, err = rt.db.GetUserById(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("user not found")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the user is part of the conversation
	isMember, err := rt.db.IsUserInConversation(userId, conversationId)
	if err != nil {
		ctx.Logger.WithError(err).Error("error checking if user is in conversation")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isMember {
		ctx.Logger.Error("user is not a member of the conversation")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Verify if the message exists
	message, err := rt.db.GetMessageById(messageId, conversationId)
	if err != nil {
		ctx.Logger.WithError(err).Error("message not found")
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Verify ownership or permissions
	if message.Sender.UserId != userId {
		ctx.Logger.Error("user unauthorized to delete message")
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Delete the message
	err = rt.db.DeleteMessageById(messageId)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to delete message")
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	// Respond with no content (204)
	w.WriteHeader(http.StatusNoContent)
}
