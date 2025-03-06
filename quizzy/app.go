package quizzy

import (
	"fmt"
	"github.com/doquangtan/socket.io/v4"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"quizzy.app/backend/quizzy/auth"
	"quizzy.app/backend/quizzy/cfg"
	quizzyhttp "quizzy.app/backend/quizzy/http"
	"quizzy.app/backend/quizzy/services"
	"time"
)

func Run() {
	config := cfg.LoadCfgFromEnv()

	// Configure GIN execution mode (dev, test, production).
	setGinMode(config.Env.AsString())

	log.Printf("application running in %s mode.\n", config.Env)

	// Initializing GIN engine.
	engine := gin.Default()
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "Connection"},
		ExposeHeaders:    []string{"Content-Length", "Location"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowWebSockets:  true,
	}))

	router := engine.Group(config.BasePath)
	router.Use(cfg.ProvideConfig(config))

	// Configure database provider.
	// Firebase access is injected here into GIN context,
	// this will enable fast access to database through handling chain itself.

	// External services must be only initialized for DEVELOPMENT or PRODUCTION modes.
	// TEST mode use dummy implementations.
	if client, err := services.ConfigureFirebase(config); err == nil {
		router.Use(func(ctx *gin.Context) {
			//FIXME: Firebase application must be initialized outside ConfigureFirebase().
			// Firestore can be initialized each time we need it.
			ctx.Set("firebase-services", client)
		})
		router.Use(auth.ProvideAuthenticator(&auth.FirebaseAuthenticator{Fbs: &client}))
	} else {
		fmt.Printf("failed to initialize firebase services: %s\n", err)
	}
	router.Use(func(ctx *gin.Context) {
		if redis := services.ConfigureRedis(config); redis != nil {
			ctx.Set("redis-service", redis)
		} else {
			fmt.Println("failed to initialize redis connection.")
		}
	})

	// Initializing SocketIO server.
	io := socketio.New()

	// Initializing HTTP routes and SocketIO.
	quizzyhttp.ConfigureRouting(router, io)
	
	// Running server...
	if err := engine.Run(config.Addr); err != nil {
		log.Fatalf("Failed to start server on %s: %s\n", config.Addr, err)
	}
}

func setGinMode(env string) {
	switch env {
	case cfg.EnvDevelopment:
		gin.SetMode(gin.DebugMode)
		break
	case cfg.EnvTest:
		gin.SetMode(gin.TestMode)
		break
	default:
		gin.SetMode(gin.ReleaseMode)
		break
	}
}
