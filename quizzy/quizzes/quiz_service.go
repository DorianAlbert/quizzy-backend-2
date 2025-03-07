package quizzes

import "errors"

var ErrQuizNotReady = errors.New("quiz not ready")

type QuizService interface {
	Create(ownerId string, quiz Quiz) error

	Get(ownerId, id string) (Quiz, error)

	GetAll(ownerId string) ([]Quiz, error)

	Patch(ownerId, quizId string, fields []FieldPatchOp) error

	CreateQuestion(ownerId string, quiz Quiz, question Question) error

	UpdateQuestion(ownerId, quizId string, question Question) error

	// StartQuiz starts the given Quiz. If the quiz doesn't meet
	// validation requirements, ErrQuizNotReady is returned.
	StartQuiz(ownerId string, quiz Quiz) error

	QuizFromCode(code string) (Quiz, error)
	IncrRoomPeople(roomId string) error
	GetRoomPeople(roomId string) (int, error)
	ResetRoomPeople(roomId string) error
}
