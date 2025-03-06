package quizzes

import (
	"encoding/json"
	"fmt"
	"sync/atomic"

	socketio "github.com/doquangtan/socket.io/v4"
	"github.com/gin-gonic/gin"
)

var participants int32

func configureIo(router *gin.RouterGroup, ws *socketio.Io) {
	ws.OnConnection(func(socket *socketio.Socket) {
		atomic.AddInt32(&participants, 1)
		updateStatus(ws)
		fmt.Println("Client connecté:", socket.Id)

		socket.On("host", func(event *socketio.EventPayload) {
			if len(event.Data) < 1 {
				fmt.Println("Pas assez de données envoyées")
				return
			}

			dataBytes, err := json.Marshal(event.Data[0])
			if err != nil {
				fmt.Printf("Erreur lors du Marshal des données : %v\n", err)
				return
			}

			var payload hostEvent
			if err := json.Unmarshal(dataBytes, &payload); err != nil {
				fmt.Printf("Erreur de parsing JSON : %v\n", err)
				return
			}

			fmt.Printf("Received quiz code: %s\n", payload.ExecutionId)

			hostResponse := hostDetailsResponse{
				Quiz: Quiz{
					Code:  "sdlmkgjdlfkmdlmgkdfl",
					Title: "Quiz Example Title",
				},
			}
			if socket == nil {
				fmt.Println("Impossible d'émettre, le socket est nil.")
				err := socket.Emit("hostDetails", hostResponse)
				if err != nil {
					fmt.Printf("Erreur lors de l'émission de 'hostDetails': %v\n", err)
				}
			} else {
				fmt.Println("Impossible d'émettre, le socket n'est pas actif.")
			}

			updateStatus(ws)
		})

		socket.On("disconnect", func(event *socketio.EventPayload) {
			atomic.AddInt32(&participants, -1)
			updateStatus(ws)
			fmt.Println("Client déconnecté:", socket.Id)
		})
	})

	router.GET("/socket.io/*any", gin.WrapH(ws.HttpHandler()))
	router.POST("/socket.io/*any", gin.WrapH(ws.HttpHandler()))
}

func updateStatus(ws *socketio.Io) {
	statusResponse := statusEvent{
		Status:       "waiting",
		Participants: fmt.Sprintf("%d", atomic.LoadInt32(&participants)),
	}
	ws.Emit("status", statusResponse)
}

type hostEvent struct {
	ExecutionId string `json:"executionId"`
}

type hostDetailsResponse struct {
	Quiz Quiz `json:"quiz"`
}

type statusEvent struct {
	Status       string `json:"status"`
	Participants string `json:"participants"`
}
