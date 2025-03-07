package quizzes

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	rooms         = make(map[string][]*websocket.Conn)
	hosts         = make(map[string]*websocket.Conn) // Stocker l'host séparément
	roomsMu       sync.Mutex
	questionIdx   = make(map[string]int) // Ajout pour suivre l'index des questions
	questionIdxMu sync.Mutex
)

// configureWs initialise la connexion WebSocket
// @Summary Connexion WebSocket
// @Description Établit une connexion WebSocket pour interagir avec le quiz en temps réel
// @Tags WebSocket
// @Produce json
// @Param Authorization header string true "Token d'authentification Bearer" default(Bearer <votre_token>)
// @Success 101 {string} string "Connexion WebSocket établie"
// @Failure 400 {string} string "Mauvaise requête"
// @Failure 401 {string} string "Non authentifié"
// @Router /quiz/ws [get]
// @Security BearerAuth
func configureWs(router *gin.RouterGroup) {
	router.GET("/", func(c *gin.Context) {
		handleWebSocket(c.Writer, c.Request, UseService(c))
	})
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, service QuizService) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to set websocket upgrade: ", err)
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		var event map[string]interface{}
		if err := json.Unmarshal(msg, &event); err != nil {
			log.Println("Error unmarshalling message:", err)
			break
		}

		switch event["name"] {
		case "host":
			handleHostEvent(conn, event["data"].(map[string]interface{}), service)
		case "join":
			handleJoinEvent(conn, event["data"].(map[string]interface{}), service)
		case "nextQuestion":
			handleNextQuestionEvent(event["data"].(map[string]interface{}), service)
		}
	}
}

// handleHostEvent permet d'héberger un quiz via WebSocket
// @Summary Héberger un quiz
// @Description L'utilisateur devient l'hôte d'un quiz et reçoit les détails du quiz
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param event body object true "Événement WebSocket 'host'"
// @Success 200 {object} map[string]interface{} "Détails du quiz et du statut"
// @Router /quiz/ws [post]
// @Security BearerAuth

func handleHostEvent(conn *websocket.Conn, data map[string]interface{}, service QuizService) {
	executionId := data["executionId"].(string)
	roomsMu.Lock()
	hosts[executionId] = conn                // On stocke l'host séparément
	rooms[executionId] = []*websocket.Conn{} // On initialise la room sans participants
	roomsMu.Unlock()

	quiz, err := service.QuizFromCode(executionId)
	if err != nil {
		return
	}
	service.ResetRoomPeople(executionId)

	response := map[string]interface{}{
		"name": "hostDetails",
		"data": map[string]interface{}{
			"quiz": quiz.Title,
		},
	}
	res, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, res)

	// Exclure l'hôte du comptage
	nbPeoples, _ := service.GetRoomPeople(executionId)

	broadcastToRoom(executionId, map[string]interface{}{
		"name": "status",
		"data": map[string]interface{}{
			"status":       "waiting",
			"participants": nbPeoples, // Pas d'incrémentation pour l'host
		},
	})
	questionIdxMu.Lock()
	questionIdx[executionId] = 0 // Initialiser l'index des questions
	questionIdxMu.Unlock()
}

// handleJoinEvent permet à un utilisateur de rejoindre un quiz via WebSocket
// @Summary Rejoindre un quiz
// @Description Un utilisateur rejoint un quiz et reçoit les détails du quiz
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param event body object true "Événement WebSocket 'join'"
// @Success 200 {object} map[string]interface{} "Détails du quiz et du statut"
// @Router /quiz/ws [post]
// @Security BearerAuth

func handleJoinEvent(conn *websocket.Conn, data map[string]interface{}, service QuizService) {
	executionId := data["executionId"].(string)
	roomsMu.Lock()
	rooms[executionId] = append(rooms[executionId], conn) // Ajout uniquement aux participants
	roomsMu.Unlock()

	quiz, err := service.QuizFromCode(executionId)
	if err != nil {
		return
	}
	service.IncrRoomPeople(executionId) // On incrémente uniquement pour les participants
	nbPeoples, _ := service.GetRoomPeople(executionId)

	response := map[string]interface{}{
		"name": "joinDetails",
		"data": map[string]interface{}{
			"quizTitle": quiz.Title,
		},
	}
	res, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, res)

	broadcastToRoom(executionId, map[string]interface{}{
		"name": "status",
		"data": map[string]interface{}{
			"status":       "waiting",
			"participants": nbPeoples, // L'hôte n'est pas compté
		},
	})
}

// handleNextQuestionEvent passe à la question suivante du quiz
// @Summary Passer à la question suivante
// @Description Envoie la prochaine question et ses réponses aux participants
// @Tags WebSocket
// @Accept json
// @Produce json
// @Param event body object true "Événement WebSocket 'nextQuestion'"
// @Success 200 {object} map[string]interface{} "Question et réponses envoyées aux participants"
// @Router /quiz/ws [post]
// @Security BearerAuth

func handleNextQuestionEvent(data map[string]interface{}, service QuizService) {
	executionId := data["executionId"].(string)
	quiz, err := service.QuizFromCode(executionId)
	if err != nil {
		return
	}

	if len(quiz.Questions) == 0 {
		return
	}

	nbPeoples, _ := service.GetRoomPeople(executionId)

	broadcastToRoom(executionId, map[string]interface{}{
		"name": "status",
		"data": map[string]interface{}{
			"status":       "started",
			"participants": nbPeoples, // Toujours sans l'hôte
		},
	})

	questionIdxMu.Lock()
	index := questionIdx[executionId]
	questionIdx[executionId]++
	questionIdxMu.Unlock()

	if index >= len(quiz.Questions) {
		return // Toutes les questions ont été posées
	}

	question := quiz.Questions[index]
	var answers []string
	for _, answer := range question.Answers {
		answers = append(answers, answer.Title)
	}

	broadcastToRoom(executionId, map[string]interface{}{
		"name": "newQuestion",
		"data": map[string]interface{}{
			"question": question.Title,
			"answers":  answers,
		},
	})
}

func broadcastToRoom(executionId string, message map[string]interface{}) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	res, _ := json.Marshal(message)

	for _, conn := range rooms[executionId] {
		conn.WriteMessage(websocket.TextMessage, res)
	}

	if host, exists := hosts[executionId]; exists {
		host.WriteMessage(websocket.TextMessage, res)
	}
}
