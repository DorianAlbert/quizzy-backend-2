package quizzes

import (
	"strings"
)

type QuizServiceImpl struct {
	store    Store
	resolver QuizCodeResolver
}

func (qs *QuizServiceImpl) Create(ownerId string, quiz Quiz) error {
	return qs.store.Upsert(ownerId, quiz)
}

func (qs *QuizServiceImpl) Get(ownerId, id string) (Quiz, error) {
	return qs.store.GetUnique(ownerId, id)
}

func (qs *QuizServiceImpl) GetAll(ownerId string) ([]Quiz, error) {
	return qs.store.GetQuizzes(ownerId)
}

func (qs *QuizServiceImpl) Patch(ownerId, quizId string, fields []FieldPatchOp) error {
	return qs.store.Patch(ownerId, quizId, fields)
}

func (qs *QuizServiceImpl) CreateQuestion(ownerId string, quiz Quiz, question Question) error {
	return qs.store.UpsertQuestion(ownerId, quiz.Id, question)
}

func (qs *QuizServiceImpl) UpdateQuestion(ownerId, quizId string, question Question) error {
	return qs.store.UpdateQuestion(ownerId, quizId, question)
}

func (qs *QuizServiceImpl) StartQuiz(ownerId string, quiz Quiz) error {
	if !quiz.Validate() {
		return ErrQuizNotReady
	}

	if err := qs.resolver.BindCode(ownerId, quiz); err != nil {
		return err
	}

	return nil
}

func (qs *QuizServiceImpl) QuizFromCode(code string) (Quiz, error) {
	if str, err := qs.resolver.GetQuiz(code); err != nil {
		return Quiz{}, err
	} else {
		ids := strings.Split(str, "@")
		return qs.store.GetUnique(ids[0], ids[1])
	}
}
func (qs *QuizServiceImpl) IncrRoomPeople(roomId string) error {
	return qs.resolver.IncrRoomPeople(roomId)
}

func (qs *QuizServiceImpl) GetRoomPeople(roomId string) (int, error) {
	return qs.resolver.GetRoomPeople(roomId)
}
