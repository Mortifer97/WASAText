package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Mortifer97/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// ForwardMessageRequest rappresenta il payload della richiesta per inoltrare un messaggio
// Handler per inoltrare un messaggio in un'altra conversazione.
// Controlla che l'utente sia membro sia della conversazione di origine che di destinazione.
// Si collega a ForwardMessage in database/message.go.
type ForwardMessageRequest struct {
	ConversationId int64 `json:"conversationId"`
}

// forwardMessage gestisce la richiesta API
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Estrae userId, conversationId e messageId dai parametri del percorso
	userIdStr := ps.ByName("userId")
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)
	conversationIdStr := ps.ByName("conversationId")
	conversationId, _ := strconv.ParseInt(conversationIdStr, 10, 64)
	messageIdStr := ps.ByName("messageId")
	messageId, _ := strconv.ParseInt(messageIdStr, 10, 64)
	if userId <= 0 || conversationId <= 0 || messageId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}

	// Controlla se l'utente esiste
	_, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}

	// Controlla se l'utente è membro della conversazione
	isMember, err := rt.db.IsUserInConversation(userId, conversationId)
	if err != nil || !isMember {
		http.Error(w, "Non autorizzato", http.StatusForbidden)
		return
	}

	// Decodifica il corpo della richiesta
	var req ForwardMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Richiesta non valida", http.StatusBadRequest)
		return
	}

	// Recupera il messaggio originale dal database
	originalMessage, err := rt.db.GetMessageById(messageId, conversationId)
	if err != nil {
		http.Error(w, "Messaggio non trovato", http.StatusNotFound)
		return
	}

	// Controlla se l'utente è membro della conversazione di destinazione
	isMember, err = rt.db.IsUserInConversation(userId, req.ConversationId)
	if err != nil || !isMember {
		http.Error(w, "Non autorizzato", http.StatusForbidden)
		return
	}

	// Inoltra il messaggio
	forwardedMessage, err := rt.db.ForwardMessage(userId, originalMessage, req.ConversationId)
	if err != nil {
		http.Error(w, "Errore inoltro messaggio", http.StatusInternalServerError)
		return
	}

	// Risponde con il messaggio inoltrato
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(forwardedMessage)
}
