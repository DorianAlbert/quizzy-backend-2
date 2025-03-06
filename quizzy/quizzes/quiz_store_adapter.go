package quizzes

import (
	"errors"
)

var (
	ErrNotFound             = errors.New("user not found")
	ErrInvalidPatchOperator = errors.New("invalid patch operator")
	ErrInvalidPatchField    = errors.New("invalid patch field")
)

type FieldPatchOp struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}
type Links struct {
	Create string `json:"create,omitempty"`
	Start  string `json:"start,omitempty"`
}

// Quiz describe available data for a quizzes.
type Quiz struct {
	Id          string     `firestore:"-" json:"id"`
	Title       string     `firestore:"title" json:"title"`
	Description string     `firestore:"description" json:"description"`
	Questions   []Question `firestore:"-" json:"questions"`
	Code        string     `firestore:"code" json:"code,omitempty"`
}

func (q *Quiz) Validate() bool {
	if len(q.Title) == 0 {
		return false
	}
	if len(q.Questions) == 0 {
		return false
	}
	for _, quest := range q.Questions {
		if !quest.Validate() {
			return false
		}
	}
	return true
}

type Question struct {
	Id      string   `firestore:"-" json:"id"`
	Title   string   `firestore:"title" json:"title"`
	Answers []Answer `firestore:"-" json:"answers"`
}

func (q *Question) Validate() bool {
	if len(q.Title) == 0 || len(q.Answers) < 2 {
		return false
	}

	validAnswers := 0

	for _, answer := range q.Answers {
		if answer.IsCorrect {
			validAnswers++
		}
	}

	if validAnswers == len(q.Answers) {
		return false
	}

	return true
}

type Answer struct {
	Id        string `firestore:"-" json:"id"`
	Title     string `firestore:"title" json:"title"`
	IsCorrect bool   `firestore:"isCorrect" json:"isCorrect"`
}

type Store interface {
	// Upsert Store or update the given user, if no user with the given id exists,
	// it will be created, otherwise it will be updated.
	Upsert(ownerId string, quiz Quiz) error

	// GetUnique returns the user matching to the given uid,
	// otherwise ErrNotFound is returned.
	GetUnique(ownerId, uid string) (Quiz, error)

	// GetQuizzes returns all quizzes owned by the given user.
	GetQuizzes(ownerId string) ([]Quiz, error)

	// Patch update the given quizzes.
	Patch(ownerId, uid string, fields []FieldPatchOp) error

	GetUniqueQuestion(ownerId, quizId, questionId string) (Question, error)

	UpsertQuestion(ownerId, quizId string, question Question) error

	// UpdateQuestion patch the given
	UpdateQuestion(ownerId, quizId string, question Question) error
}

type QuizCodeResolver interface {
	BindCode(ownerId string, quiz Quiz) error
	UnbindCode(code string) error
	GetQuiz(code string) (string, error)
	IncrRoomPeople(roomId string) error
	GetRoomPeople(roomId string) (int, error)
}
