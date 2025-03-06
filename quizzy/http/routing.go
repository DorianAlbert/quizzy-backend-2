package http

import (
	"github.com/doquangtan/socket.io/v4"
	"github.com/gin-gonic/gin"
	"quizzy.app/backend/quizzy/ping"
	"quizzy.app/backend/quizzy/quizzes"
	"quizzy.app/backend/quizzy/users"
)

func ConfigureRouting(router *gin.RouterGroup, io *socketio.Io) {
	ping.ConfigureRoutes(router)
	users.ConfigureRoutes(router)
	quizzes.ConfigureRoutes(router, io)
}
