package api

import (
	"net/http"
	"strconv"

	"github.com/Mortifer97/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Registrazione delle rotte API principali.
	// Ogni rotta è associata a una funzione handler che gestisce la richiesta.
	// Le funzioni AuthHandler garantiscono che l'utente sia autenticato prima di proseguire.
	// Alcune rotte si collegano a funzioni di altri file, come la gestione dei messaggi, gruppi e utenti.
	// I dettagli delle funzioni sono definiti nei rispettivi file handler.
	//
	// Esempio: la rotta POST /users/:userId/conversations/:conversationId/messages/ richiama postMessage,
	// che a sua volta interagisce con il database tramite le funzioni in /service/database/message.go.

	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	rt.router.POST("/session", rt.wrap(rt.postSession))
	rt.router.PUT("/users/:userId/username", rt.wrap(rt.AuthHandler(rt.putUsername)))
	rt.router.GET("/users/:userId/conversations/", rt.wrap(rt.AuthHandler(rt.getConversations)))
	rt.router.PUT("/users/:userId/conversations/", rt.wrap(rt.AuthHandler(rt.addConversation)))
	rt.router.GET("/users/:userId/conversations/:conversationId", rt.wrap(rt.AuthHandler(rt.getConversation)))
	rt.router.POST("/users/:userId/conversations/:conversationId/messages/", rt.wrap(rt.AuthHandler(rt.postMessage)))
	rt.router.POST("/users/:userId/conversations/:conversationId/messages/:messageId/forwardMessage", rt.wrap(rt.AuthHandler(rt.forwardMessage)))
	rt.router.POST("/users/:userId/conversations/:conversationId/messages/:messageId/replyMessage", rt.wrap(rt.AuthHandler(rt.replyMessage)))
	rt.router.PUT("/users/:userId/conversations/:conversationId/messages/:messageId/comments/", rt.wrap(rt.AuthHandler(rt.commentMessage)))
	rt.router.DELETE("/users/:userId/conversations/:conversationId/messages/:messageId/comments/:commentId", rt.wrap(rt.AuthHandler(rt.removeComment)))
	rt.router.DELETE("/users/:userId/conversations/:conversationId/messages/:messageId", rt.wrap(rt.AuthHandler(rt.deleteMessage)))
	rt.router.PUT("/users/:userId/groups/:groupId/members/", rt.wrap(rt.AuthHandler(rt.addToGroup)))
	rt.router.DELETE("/users/:userId/groups/:groupId/members/me", rt.wrap(rt.AuthHandler(rt.leaveGroup)))
	rt.router.PUT("/users/:userId/groups/:groupId/name", rt.wrap(rt.AuthHandler(rt.setGroupName)))
	rt.router.PUT("/users/:userId/photo", rt.wrap(rt.AuthHandler(rt.setMyPhoto)))
	rt.router.PUT("/users/:userId/groups/:groupId/photo", rt.wrap(rt.AuthHandler(rt.setGroupPhoto)))
	rt.router.GET("/users/:userId/search", rt.wrap(rt.AuthHandler(rt.searchUsers)))
	rt.router.GET("/users/:userId/groups/:groupId/members/", rt.wrap(rt.AuthHandler(rt.getGroupMembers)))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}

// Handler per autenticazione utente tramite header Authorization.
// Se l'header manca o l'ID non è valido, restituisce errore 401.
// Se l'utente esiste, passa il controllo all'handler successivo.
// Si collega a tutte le rotte che richiedono autenticazione (vedi api-handler.go).
func (rt *_router) AuthHandler(next httpRouterHandler) httpRouterHandler {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			ctx.Logger.Warn("header Authorization mancante")
			http.Error(w, "Non autorizzato: header mancante", http.StatusUnauthorized)
			return
		}
		userID, err := strconv.ParseInt(authHeader, 10, 64)
		if err != nil {
			ctx.Logger.Warn("formato userID non valido")
			http.Error(w, "Non autorizzato: formato userID non valido", http.StatusUnauthorized)
			return
		}
		_, err = rt.db.GetUserById(userID)
		if err != nil {
			ctx.Logger.Warn("userID non valido")
			http.Error(w, "Non autorizzato: userID non valido", http.StatusUnauthorized)
			return
		}
		next(w, r, ps, ctx)
	}
}
