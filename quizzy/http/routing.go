package http

import (
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"quizzy.app/backend/quizzy/ping"
	"quizzy.app/backend/quizzy/quizzes"
	"quizzy.app/backend/quizzy/users"
)

func ConfigureRouting(router *gin.RouterGroup, ws *socketio.Server) {
	ping.ConfigureRoutes(router)
	users.ConfigureRoutes(router)
	quizzes.ConfigureRoutes(router, ws)
}
