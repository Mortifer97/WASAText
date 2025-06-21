package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Mortifer97/WASAText/service/api/reqcontext"
	"github.com/Mortifer97/WASAText/service/database"
	"github.com/julienschmidt/httprouter"
)

// Handler per ottenere i messaggi di una conversazione.
// Controlla che l'utente esista e recupera i messaggi tramite GetMessagesByConversation.
// Si collega a database/message.go.
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Estrae e converte gli ID utente e conversazione dai parametri del percorso
	userId, _ := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	conversationId, _ := strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if userId <= 0 || conversationId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}

	// Estrae e convalida il parametro di query "sort"
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	// Controlla se l'utente esiste
	_, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}

	// Recupera i messaggi dal database
	messages, err := rt.db.GetMessagesByConversation(userId, conversationId, sortOrder)
	if err != nil {
		http.Error(w, "Errore recupero messaggi", http.StatusInternalServerError)
		return
	}

	// Payload della risposta
	response := struct {
		ConversationID int64              `json:"conversationId"`
		Messages       []database.Message `json:"messages"`
	}{
		ConversationID: conversationId,
		Messages:       messages,
	}

	// Risponde con i dettagli della conversazione
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handler per ottenere la lista delle conversazioni di un utente.
// Valida il parametro sort e chiama GetConversationsByUser.
// Si collega a database/conversation.go.
func (rt *_router) getConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, _ := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if userId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	_, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}
	conversations, err := rt.db.GetConversationsByUser(userId, sortOrder)
	if err != nil {
		http.Error(w, "Errore recupero conversazioni", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// Handler per ottenere i membri di un gruppo.
// Controlla che l'utente sia membro del gruppo e restituisce la lista tramite GetGroupMembers.
// Si collega a database/group.go.
func (rt *_router) getGroupMembers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, _ := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	groupId, _ := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if userId <= 0 || groupId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}
	isMember, err := rt.db.IsUserMemberOfGroup(userId, groupId)
	if err != nil || !isMember {
		http.Error(w, "Non autorizzato", http.StatusForbidden)
		return
	}
	members, err := rt.db.GetGroupMembers(groupId)
	if err != nil {
		http.Error(w, "Errore recupero membri", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// Handler per la ricerca utenti tramite username.
// Controlla che l'utente esista e chiama SearchUsersByUsername.
// Si collega a database/user.go.
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Estrae l'ID utente dai parametri del percorso
	userIdStr := ps.ByName("userId")
	userId, errUsr := strconv.ParseInt(userIdStr, 10, 64)
	if errUsr != nil || userId <= 0 {
		ctx.Logger.WithError(errUsr).Error("ID utente non valido")
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}

	// Estrae il parametro di query "username"
	username := r.URL.Query().Get("username")

	// Controlla se l'utente esiste
	_, err := rt.db.GetUserById(userId)
	if err != nil {
		ctx.Logger.WithError(err).Error("utente non trovato")
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}

	// Esegue la ricerca nel database per username
	users, err := rt.db.SearchUsersByUsername(username)
	if err != nil {
		ctx.Logger.WithError(err).Error("errore nella ricerca utenti nel database")
		http.Error(w, "Errore ricerca utenti", http.StatusInternalServerError)
		return
	}

	// Restituisce l'elenco degli utenti corrispondenti
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		ctx.Logger.WithError(err).Error("errore nella codifica della risposta")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
