package quizzes

import (
	"errors"
	"fmt"
	"github.com/doquangtan/socket.io/v4"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"quizzy.app/backend/quizzy/auth"
)

func ConfigureRoutes(rt *gin.RouterGroup, io *socketio.Io) {
	// Settings up SocketIO configuration for quiz module.
	configureIo(rt, io)

	secured := rt.Group("/quiz", auth.RequireAuthenticated, ProvideService)
	secured.GET("", handleGetAllUserQuiz)
	secured.POST("", handlePostQuiz)

	quiz := secured.Group("/:quiz-id", ProvideQuiz)
	quiz.GET("", handleGetQuiz)
	quiz.PATCH("", handlePatchQuiz)
	quiz.GET("/questions", handleGetQuestions)
	quiz.POST("/questions", handlePostQuestion)
	quiz.PUT("/questions/:question-id", ProvideQuestion, handlePutQuestion)
	quiz.POST("/start", handleStartQuiz)
}

func handleGetQuiz(ctx *gin.Context) {
	quiz := UseQuiz(ctx)
	ctx.JSON(http.StatusOK, quiz)
}

type QuizWithLinks struct {
	Quiz
	Links Links `json:"_links"`
}

type UserQuizzesResponse struct {
	Data  []QuizWithLinks `json:"data"`
	Links Links           `json:"_links"`
}

func mapMultipleQuizWithLinks(quizzes []Quiz) []QuizWithLinks {
	qwl := make([]QuizWithLinks, len(quizzes))

	for i := range quizzes {
		qwl[i] = mapQuizWithLinks(quizzes[i])
	}

	return qwl
}

func mapQuizWithLinks(quiz Quiz) QuizWithLinks {
	lnk := Links{
		Create: "",
		Start:  "",
	}

	if quiz.Validate() {
		lnk.Start = fmt.Sprintf("http://localhost:8000/quiz/%s/start", quiz.Id)
	}

	return QuizWithLinks{
		Quiz:  quiz,
		Links: lnk,
	}
}

func handleGetAllUserQuiz(ctx *gin.Context) {
	id := auth.UseIdentity(ctx)
	service := UseService(ctx)

	if quizzes, err := service.GetAll(id.Uid); err == nil {
		ctx.JSON(http.StatusOK, UserQuizzesResponse{
			Data: mapMultipleQuizWithLinks(quizzes),
			Links: Links{
				Create: "http://localhost:8000/quiz",
			},
		})
		return
	}

	ctx.AbortWithStatus(http.StatusInternalServerError)
}

type CreateQuizRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func handlePostQuiz(ctx *gin.Context) {
	id := auth.UseIdentity(ctx)
	service := UseService(ctx)

	// Getting payload from request.
	var req CreateQuizRequest
	if ctx.ShouldBindJSON(&req) != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if code, err := GenerateCode(); err == nil {
		quiz := Quiz{
			Id:          uuid.New().String(),
			Title:       req.Title,
			Description: req.Description,
			Code:        code,
		}

		if err2 := service.Create(id.Uid, quiz); err2 == nil {
			ctx.Header("Location", fmt.Sprintf("http://localhost:8000/quiz/%s", quiz.Id))
			ctx.JSON(http.StatusCreated, quiz)
			return
		}
	}

	// WARNING / WARNING / WARNING //
	// If this code-path is reached, it means that the requested user was never
	// registered in our firestore.
	// WARNING / WARNING / WARNING //

	// This may happen if client does some weird things...
	// We should never let the client decide about this process,
	// user registration must be done in single request (Client->Server), or we must use pub/sub (or something similar)
	// from firebase to service user automatically to avoid data consistency issues.
	ctx.AbortWithStatus(http.StatusInternalServerError)
}

type PatchQuizRequest []FieldPatchOp

func handlePatchQuiz(ctx *gin.Context) {
	id := auth.UseIdentity(ctx)
	store := UseService(ctx)
	quiz := UseQuiz(ctx)

	var req PatchQuizRequest
	if ctx.ShouldBindJSON(&req) != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := store.Patch(id.Uid, quiz.Id, req); err == nil {
		ctx.Status(http.StatusNoContent)
	} else if errors.Is(err, ErrInvalidPatchOperator) || errors.Is(err, ErrInvalidPatchField) {
		ctx.AbortWithStatus(http.StatusBadRequest)
	} else {
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}

type CreateQuestionRequest struct {
	Title   string   `json:"title"`
	Answers []Answer `json:"answers"`
}

func handlePostQuestion(ctx *gin.Context) {
	service := UseService(ctx)
	id := auth.UseIdentity(ctx)
	quiz := UseQuiz(ctx)

	var req CreateQuestionRequest
	if ctx.ShouldBindJSON(&req) != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	question := Question{
		Id:      uuid.New().String(),
		Title:   req.Title,
		Answers: req.Answers,
	}
	err := service.CreateQuestion(id.Uid, quiz, question)

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Header("Location", fmt.Sprintf("http://localhost:8000/quiz/%s/questions/%s", quiz.Id, question.Id))
	ctx.Status(http.StatusCreated)
}

func handleGetQuestions(ctx *gin.Context) {
	quiz := UseQuiz(ctx)
	ctx.JSON(http.StatusOK, quiz.Questions)
}

type UnidentifiedAnswer struct {
	Title     string `json:"title"`
	IsCorrect bool   `json:"isCorrect"`
}

type UpdateQuestionRequest struct {
	Title   string               `json:"title"`
	Answers []UnidentifiedAnswer `json:"answers"`
}

func handlePutQuestion(ctx *gin.Context) {
	id := auth.UseIdentity(ctx)
	store := UseService(ctx)
	quiz := UseQuiz(ctx)
	question := UseQuestion(ctx)

	var payload UpdateQuestionRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	question.Title = payload.Title
	question.Answers = make([]Answer, 0)
	for _, a := range payload.Answers {
		question.Answers = append(question.Answers, Answer{
			Id:        uuid.New().String(),
			Title:     a.Title,
			IsCorrect: a.IsCorrect,
		})
	}

	if err := store.UpdateQuestion(id.Uid, quiz.Id, question); err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func handleStartQuiz(ctx *gin.Context) {
	service := UseService(ctx)
	identity := auth.UseIdentity(ctx)
	quiz := UseQuiz(ctx)

	if err := service.StartQuiz(identity.Uid, quiz); errors.Is(err, ErrQuizNotReady) {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	} else if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ctx.Header("Location", fmt.Sprintf("http://localhost:8000/execution/%s", quiz.Code))
	ctx.Status(http.StatusCreated)
}
