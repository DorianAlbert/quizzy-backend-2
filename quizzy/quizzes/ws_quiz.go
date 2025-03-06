package quizzes

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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
			handleNextQuestionEvent(conn, event["data"].(map[string]interface{}), service)
		}
	}
}

func handleHostEvent(conn *websocket.Conn, data map[string]interface{}, service QuizService) {
	executionId := data["executionId"].(string)
	quiz, err := service.QuizFromCode(executionId)
	if err != nil {
		return
	}
	service.IncrRoomPeople(executionId)
	nbPeoples, err := service.GetRoomPeople(executionId)
	if err != nil {
		return
	}

	// Simuler une réponse
	response := map[string]interface{}{
		"name": "hostDetails",
		"data": map[string]interface{}{
			"quiz": quiz.Title,
		},
	}
	res, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, res)
	response2 := map[string]interface{}{
		"name": "status",
		"data": map[string]any{
			"status":       "waiting",
			"participants": nbPeoples,
		},
	}
	res2, _ := json.Marshal(response2)
	conn.WriteMessage(websocket.TextMessage, res2)
}

func handleJoinEvent(conn *websocket.Conn, data map[string]interface{}, service QuizService) {
	executionId := data["executionId"].(string)
	quiz, err := service.QuizFromCode(executionId)
	if err != nil {
		return
	}

	fmt.Println("ExecutionId reçu:", executionId)

	service.IncrRoomPeople(executionId)
	nbPeoples, err := service.GetRoomPeople(executionId)
	if err != nil {
		return
	}

	// Simuler une réponse
	response := map[string]interface{}{
		"name": "joinDetails",
		"data": map[string]interface{}{
			"quizTitle": quiz.Title,
		},
	}
	res, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, res)

	// Émettre un événement de statut
	statusResponse := map[string]interface{}{
		"name": "status",
		"data": map[string]interface{}{
			"status":       "waiting",
			"participants": nbPeoples,
		},
	}
	statusRes, _ := json.Marshal(statusResponse)
	conn.WriteMessage(websocket.TextMessage, statusRes)
}

func handleNextQuestionEvent(conn *websocket.Conn, data map[string]interface{}, service QuizService) {
	executionId := data["executionId"].(string)
	fmt.Println("ExecutionId reçu:", executionId)
	service.IncrRoomPeople(executionId)
	nbPeoples, err := service.GetRoomPeople(executionId)
	if err != nil {
		return
	}
	quiz, err := service.QuizFromCode(executionId)
	if err != nil {
		return
	}

	// Émettre un événement de statut
	statusResponse := map[string]interface{}{
		"name": "status",
		"data": map[string]interface{}{
			"status":       "started",
			"participants": nbPeoples,
		},
	}
	statusRes, _ := json.Marshal(statusResponse)
	conn.WriteMessage(websocket.TextMessage, statusRes)

	// Vérifier qu'il y a des questions dans le quiz
	if len(quiz.Questions) == 0 {
		return
	}

	// Récupérer la première question
	question := quiz.Questions[0]

	// Extraire les réponses
	var answers []string
	for _, answer := range question.Answers {
		answers = append(answers, answer.Title)
	}

	// Émettre une nouvelle question
	questionResponse := map[string]interface{}{
		"name": "newQuestion",
		"data": map[string]interface{}{
			"question": question.Title,
			"answers":  answers,
		},
	}
	questionRes, _ := json.Marshal(questionResponse)
	conn.WriteMessage(websocket.TextMessage, questionRes)
}
