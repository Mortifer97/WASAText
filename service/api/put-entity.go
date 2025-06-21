package api

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"unicode/utf8"

	"github.com/Mortifer97/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// CommentRequest rappresenta il payload della richiesta per aggiungere un commento a un messaggio
// Handler per aggiungere un commento (emoji) a un messaggio.
// Controlla che l'utente sia membro della conversazione e chiama AddCommentToMessage.
// Si collega a database/comment.go.
type CommentRequest struct {
	Content string `json:"content"`
}

func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
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
	var commentRequest CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&commentRequest); err != nil {
		http.Error(w, "Richiesta non valida", http.StatusBadRequest)
		return
	}
	if !isValidEmoji(commentRequest.Content) {
		http.Error(w, "Formato emoji non valido", http.StatusBadRequest)
		return
	}
	_, err = rt.db.GetMessageById(messageId, conversationId)
	if err != nil {
		http.Error(w, "Messaggio non trovato", http.StatusNotFound)
		return
	}
	newComment, err := rt.db.AddCommentToMessage(messageId, userId, commentRequest.Content)
	if err != nil {
		http.Error(w, "Errore aggiunta commento", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newComment)
}

// isValidEmoji controlla se il contenuto è una reazione emoji valida
func isValidEmoji(emoji string) bool {
	r, _ := utf8.DecodeRuneInString(emoji)
	if (r >= '\U0001F600' && r <= '\U0001F64F') ||
		(r >= '\U0001F300' && r <= '\U0001F5FF') ||
		(r >= '\U0001F680' && r <= '\U0001F6FF') ||
		(r >= '\U0001F700' && r <= '\U0001F77F') {
		return true
	}
	return false
}

// Handler per avviare una nuova conversazione (diretta o gruppo).
// Controlla che l'utente e il target esistano, poi chiama CreateConversation.
// Si collega a database/conversation.go.
func (rt *_router) addConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, _ := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if userId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}
	user, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}
	var body struct {
		TargetUsername string `json:"targetUsername"`
		Type           string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Richiesta non valida", http.StatusBadRequest)
		return
	}
	targetUser, err := rt.db.GetUserByName(body.TargetUsername)
	if err != nil {
		http.Error(w, "Utente target non trovato", http.StatusNotFound)
		return
	}
	if body.Type != "group" && body.Type != "direct" {
		http.Error(w, "Tipo conversazione non valido", http.StatusBadRequest)
		return
	}
	conversation, err := rt.db.CreateConversation(user.UserId, targetUser.UserId, body.Type)
	if err != nil {
		http.Error(w, "Errore creazione conversazione", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conversation)
}

// Handler per cambiare il nome di un gruppo.
// Controlla che l'utente sia membro e aggiorna il nome tramite UpdateGroupName.
// Si collega a database/group.go.
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, _ := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	groupId, _ := strconv.Atoi(ps.ByName("groupId"))
	if userId <= 0 || groupId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}
	var requestBody struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Richiesta non valida", http.StatusBadRequest)
		return
	}
	if len(requestBody.Name) < 1 || len(requestBody.Name) > 32 || !isValidGroupName(requestBody.Name) {
		http.Error(w, "Nome gruppo non valido", http.StatusBadRequest)
		return
	}
	_, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}
	group, err := rt.db.GetGroupById(int64(groupId))
	if err != nil {
		http.Error(w, "Gruppo non trovato", http.StatusNotFound)
		return
	}
	isMember, err := rt.db.IsUserMemberOfGroup(userId, int64(groupId))
	if err != nil || !isMember {
		http.Error(w, "Non sei membro del gruppo", http.StatusNotFound)
		return
	}
	err = rt.db.UpdateGroupName(int64(groupId), requestBody.Name)
	if err != nil {
		http.Error(w, "Errore aggiornamento nome gruppo", http.StatusInternalServerError)
		return
	}
	response := struct {
		GroupId int64  `json:"groupId"`
		NewName string `json:"newName"`
	}{
		GroupId: group.ConversationId,
		NewName: requestBody.Name,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// isValidGroupName validates the group name (only allowing alphanumeric characters and spaces)
func isValidGroupName(name string) bool {
	// Regular expression to match valid group names
	re := regexp.MustCompile(`^[a-zA-Z0-9 ]+$`)
	return re.MatchString(name)
}

// Handler per aggiornare la foto di un gruppo.
// Controlla che l'utente sia membro e aggiorna la foto tramite UpdateGroupPhoto.
// Si collega a database/image.go.
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, _ := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	groupId, _ := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if userId <= 0 || groupId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}
	_, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}
	group, err := rt.db.GetGroupById(int64(groupId))
	if err != nil {
		http.Error(w, "Gruppo non trovato", http.StatusNotFound)
		return
	}
	isMember, err := rt.db.IsUserMemberOfGroup(userId, int64(groupId))
	if err != nil || !isMember {
		http.Error(w, "Non sei membro del gruppo", http.StatusNotFound)
		return
	}
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Errore parsing form", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Foto non valida", http.StatusBadRequest)
		return
	}
	defer file.Close()
	photoData, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Errore lettura foto", http.StatusInternalServerError)
		return
	}
	err = rt.db.UpdateGroupPhoto(group.ConversationId, photoData)
	if err != nil {
		http.Error(w, "Errore aggiornamento foto gruppo", http.StatusInternalServerError)
		return
	}
	response := struct {
		GroupId int64  `json:"groupId"`
		Photo   []byte `json:"photo,omitempty"`
	}{
		GroupId: groupId,
		Photo:   photoData,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Handler per aggiornare la foto dell'utente.
// Controlla che l'utente esista e aggiorna la foto tramite UpdateUserPhoto.
// Si collega a database/image.go.
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, _ := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	if userId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}
	user, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Errore parsing form", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Foto non valida", http.StatusBadRequest)
		return
	}
	defer file.Close()
	photoData, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Errore lettura foto", http.StatusInternalServerError)
		return
	}
	err = rt.db.UpdateUserPhoto(user.UserId, photoData)
	if err != nil {
		http.Error(w, "Errore aggiornamento foto", http.StatusInternalServerError)
		return
	}
	response := struct {
		UserId int64  `json:"userId"`
		Photo  string `json:"photo,omitempty"`
	}{
		UserId: userId,
		Photo:  "",
	}
	if len(photoData) > 0 {
		response.Photo = base64.StdEncoding.EncodeToString(photoData)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Handler per aggiungere un utente a un gruppo.
// Controlla che l'utente e il gruppo esistano, poi chiama AddUserToGroup.
// Si collega a database/group.go.
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId, _ := strconv.ParseInt(ps.ByName("userId"), 10, 64)
	groupId, _ := strconv.Atoi(ps.ByName("groupId"))
	if userId <= 0 || groupId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}
	_, err := rt.db.GetUserById(userId)
	if err != nil {
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}
	var body struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Richiesta non valida", http.StatusBadRequest)
		return
	}
	existingUser, err := rt.db.GetUserByName(body.Username)
	if err != nil {
		http.Error(w, "Utente non trovato", http.StatusNotFound)
		return
	}
	_, err = rt.db.GetGroupById(int64(groupId))
	if err != nil {
		http.Error(w, "Gruppo non trovato", http.StatusNotFound)
		return
	}
	if err := rt.db.AddUserToGroup(int64(groupId), existingUser.UserId); err != nil {
		http.Error(w, "Errore aggiunta utente", http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"userId":  existingUser.UserId,
		"groupId": groupId,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UsernameRequestBody rappresenta il payload ricevuto nella richiesta
// UsernameResponse rappresenta la risposta restituita dall'endpoint
type UsernameRequestBody struct {
	Username string `json:"username"`
}
type UsernameResponse struct {
	UserID      int64  `json:"userId"`
	NewUsername string `json:"newUsername"`
}

// putUsername gestisce la richiesta API
// Handler per cambiare username dell'utente.
// Controlla che il nuovo username sia valido e non già usato, poi aggiorna tramite UpdateUsername.
// Si collega a database/user.go.
func (rt *_router) putUsername(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Estrai userId dai parametri del percorso
	userIdStr := ps.ByName("userId")
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)
	if userId <= 0 {
		http.Error(w, "Id non valido", http.StatusBadRequest)
		return
	}

	// Analizza il corpo della richiesta
	var body UsernameRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Richiesta non valida", http.StatusBadRequest)
		return
	}

	// Valida il nuovo username
	if len(body.Username) < 3 || len(body.Username) > 16 {
		http.Error(w, "Username non valido", http.StatusBadRequest)
		return
	}

	// Verifica se il nome utente esiste già
	existingUser, err := rt.db.GetUserByName(body.Username)
	if err != nil {
		http.Error(w, "Errore ricerca username", http.StatusInternalServerError)
		return
	}

	// Se l'utente esiste, restituisci un errore
	if existingUser != nil {
		http.Error(w, "Username già in uso", http.StatusBadRequest)
		return
	}

	// Aggiorna il nome dell'utente nel database
	err = rt.db.UpdateUsername(userId, body.Username)
	if err != nil {
		http.Error(w, "Errore aggiornamento username", http.StatusInternalServerError)
		return
	}

	// Costruisci la risposta
	response := UsernameResponse{
		UserID:      userId,
		NewUsername: body.Username,
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
