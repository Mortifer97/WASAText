package api

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/Mortifer97/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// Handler per la creazione di una nuova sessione utente.
// Valida il nome e crea l'utente se non esiste già.
// Si collega a CreateUser e GetUserByName in database/user.go.
type UserRequestBody struct {
	Name string `json:"name"`
}
type UserResponse struct {
	Id int64 `json:"id"`
}

// postSession gestisce la richiesta API
func (rt *_router) postSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var body UserRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Richiesta non valida", http.StatusBadRequest)
		return
	}
	if len(body.Name) < 3 || len(body.Name) > 16 {
		http.Error(w, "Nome non valido", http.StatusBadRequest)
		return
	}
	pattern := `^.*?$`
	matched, err := regexp.MatchString(pattern, body.Name)
	if err != nil || !matched {
		http.Error(w, "Nome non conforme", http.StatusBadRequest)
		return
	}
	existingUser, err := rt.db.GetUserByName(body.Name)
	if err != nil {
		http.Error(w, "Errore ricerca utente", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		response := UserResponse{Id: existingUser.UserId}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}
	newUser, err := rt.db.CreateUser(body.Name)
	if err != nil {
		http.Error(w, "Errore creazione utente", http.StatusInternalServerError)
		return
	}
	response := UserResponse{Id: newUser.UserId}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Handler per inviare un nuovo messaggio (testo o foto).
// Controlla che l'utente sia membro della conversazione e chiama AddMessage.
// Si collega a database/message.go.
func (rt *_router) postMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, _ := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	conversationId, _ := strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if userId <= 0 || conversationId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}
	_, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}
	isMember, err := rt.db.IsUserInConversation(userId, conversationId)
	if err != nil || !isMember {
		http.Error(w, "Non autorizzato", http.StatusForbidden)
		return
	}
	content := r.FormValue("content")
	if content == "" {
		file, _, err := r.FormFile("content")
		if err != nil {
			http.Error(w, "Foto non valida", http.StatusBadRequest)
			return
		}
		defer file.Close()
		photoBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Errore lettura foto", http.StatusInternalServerError)
			return
		}
		newMessage, err := rt.db.AddMessage(conversationId, userId, "", "received", "photo", photoBytes)
		if err != nil {
			http.Error(w, "Errore salvataggio foto", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newMessage)
		return
	}
	if content != "" {
		newMessage, err := rt.db.AddMessage(conversationId, userId, content, "received", "text", nil)
		if err != nil {
			http.Error(w, "Errore salvataggio messaggio", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newMessage)
	}
}

// Handler per rispondere a un messaggio (testo o foto).
// Controlla che l'utente sia membro della conversazione e chiama ReplyMessage.
// Si collega a database/message.go.
func (rt *_router) replyMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, _ := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	conversationId, _ := strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	messageId, _ := strconv.ParseInt(ps.ByName("messageId"), 10, 64)
	if userId <= 0 || conversationId <= 0 || messageId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}
	_, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}
	isMember, err := rt.db.IsUserInConversation(userId, conversationId)
	if err != nil || !isMember {
		http.Error(w, "Non autorizzato", http.StatusForbidden)
		return
	}
	content := r.FormValue("content")
	if content == "" {
		file, _, err := r.FormFile("content")
		if err != nil {
			http.Error(w, "Foto non valida", http.StatusBadRequest)
			return
		}
		defer file.Close()
		photoBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Errore lettura foto", http.StatusInternalServerError)
			return
		}
		newMessage, err := rt.db.ReplyMessage(conversationId, userId, messageId, "", "received", "photo", photoBytes)
		if err != nil {
			http.Error(w, "Errore salvataggio risposta", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newMessage)
		return
	}
	if content != "" {
		newMessage, err := rt.db.ReplyMessage(conversationId, userId, messageId, content, "received", "text", nil)
		if err != nil {
			http.Error(w, "Errore salvataggio risposta", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newMessage)
	}
}
