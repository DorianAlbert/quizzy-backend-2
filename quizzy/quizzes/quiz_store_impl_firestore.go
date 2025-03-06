package quizzes

import (
	"cloud.google.com/go/firestore"
	"context"
	"strings"
)

type quizFirestore struct {
	client *firestore.Client
}

func ConfigureStore(client *firestore.Client) Store {
	return &quizFirestore{client: client}
}

func (fs *quizFirestore) Upsert(ownerId string, quiz Quiz) error {
	_, err := fs.client.
		Doc(strings.Join([]string{"users", ownerId, "quizzes", quiz.Id}, "/")).
		Set(context.Background(), quiz)
	return err
}

func (fs *quizFirestore) GetUnique(ownerId, uid string) (Quiz, error) {
	doc, err := fs.client.
		Doc(strings.Join([]string{"users", ownerId, "quizzes", uid}, "/")).
		Get(context.Background())

	if err != nil {
		return Quiz{}, err
	}

	if !doc.Exists() {
		return Quiz{}, ErrNotFound
	}

	var quiz Quiz
	if err2 := doc.DataTo(&quiz); err2 != nil {
		return quiz, err2
	}

	quiz.Id = doc.Ref.ID

	if qs, err2 := fs.getQuestions(ownerId, quiz.Id); err2 != nil {
		return quiz, err2
	} else {
		quiz.Questions = qs
	}

	return quiz, nil
}

func (fs *quizFirestore) GetQuizzes(ownerId string) ([]Quiz, error) {
	docsIter, err := fs.client.
		Collection(strings.Join([]string{"users", ownerId, "quizzes"}, "/")).
		Documents(context.Background()).
		GetAll()

	if err != nil {
		return nil, err
	}

	// Must always be initialized to avoid nil pointer.
	arr := make([]Quiz, 0)
	for _, doc := range docsIter {
		var quiz Quiz
		if err2 := doc.DataTo(&quiz); err2 != nil {
			return nil, err2
		}

		quiz.Id = doc.Ref.ID

		if questions, err3 := fs.getQuestions(ownerId, quiz.Id); err3 != nil {
			return nil, err3
		} else {
			quiz.Questions = questions
		}

		arr = append(arr, quiz)
	}

	return arr, nil
}

func (fs *quizFirestore) Patch(ownerId, uid string, fields []FieldPatchOp) error {
	var updates []firestore.Update
	for _, op := range fields {
		if op.Op != "replace" {
			return ErrInvalidPatchOperator
		}

		// Removing unwanted leading '/'.
		// Field path for firestore must not contain anyone of : ~/*[]
		path := strings.TrimLeft(op.Path, "/")

		if strings.ContainsAny(path, "*[]~") {
			return ErrInvalidPatchField
		}

		updates = append(updates, firestore.Update{
			FieldPath: strings.Split(path, "/"),
			Value:     op.Value,
		})
	}

	_, err := fs.client.
		Collection("users").
		Doc(ownerId).
		Collection("quizzes").
		Doc(uid).
		Update(context.Background(), updates)
	return err
}

func (fs *quizFirestore) getQuestions(ownerId, quizId string) ([]Question, error) {
	docsIter, err := fs.client.
		Collection(strings.Join([]string{"users", ownerId, "quizzes", quizId, "questions"}, "/")).
		Documents(context.Background()).
		GetAll()

	if err != nil {
		return nil, err
	}

	// Must always be initialized to avoid nil pointer.
	arr := make([]Question, 0)
	for _, doc := range docsIter {
		var question Question
		if err2 := doc.DataTo(&question); err2 != nil {
			return nil, err2
		}

		question.Id = doc.Ref.ID

		if answers, err3 := fs.getAnswers(ownerId, quizId, doc.Ref.ID); err3 != nil {
			return nil, err3
		} else {
			question.Answers = answers
		}

		arr = append(arr, question)
	}

	return arr, nil
}

func (fs *quizFirestore) getAnswers(ownerId, quizId, questionId string) ([]Answer, error) {
	docsIter, err := fs.client.
		Collection(strings.Join([]string{"users", ownerId, "quizzes", quizId, "questions", questionId, "answers"}, "/")).
		Documents(context.Background()).
		GetAll()

	if err != nil {
		return nil, err
	}

	// Must always be initialized to avoid nil pointer.
	arr := make([]Answer, 0)
	for _, doc := range docsIter {
		var answer Answer
		if err2 := doc.DataTo(&answer); err2 != nil {
			return nil, err2
		}

		answer.Id = doc.Ref.ID
		arr = append(arr, answer)
	}

	return arr, nil
}

func (fs *quizFirestore) UpsertQuestion(ownerId, quizId string, question Question) error {
	if err := fs.UpdateQuestion(ownerId, quizId, question); err != nil {
		return err
	}

	return nil
}

func (fs *quizFirestore) GetUniqueQuestion(ownerId, quizId, questionId string) (Question, error) {
	doc, err := fs.client.
		Doc(strings.Join([]string{"users", ownerId, "quizzes", quizId, "questions", questionId}, "/")).
		Get(context.Background())

	if err != nil {
		return Question{}, err
	}

	if !doc.Exists() {
		return Question{}, ErrNotFound
	}

	var question Question
	if err2 := doc.DataTo(&question); err2 != nil {
		return question, err2
	}

	question.Id = doc.Ref.ID
	return question, nil
}

func (fs *quizFirestore) UpdateQuestion(ownerId, quizId string, question Question) error {
	_, err := fs.client.
		Doc(strings.Join([]string{"users", ownerId, "quizzes", quizId, "questions", question.Id}, "/")).
		Set(context.Background(), question)

	if err != nil {
		return err
	}

	iter, err2 := fs.client.
		Collection(strings.Join([]string{"users", ownerId, "quizzes", quizId, "questions", question.Id, "answers"}, "/")).
		Documents(context.Background()).
		GetAll()

	if err2 != nil {
		return err2
	}

	for _, doc := range iter {
		if _, err3 := doc.Ref.Delete(context.Background()); err3 != nil {
			return err3
		}
	}

	for _, answer := range question.Answers {
		_, err4 := fs.client.
			Doc(strings.Join([]string{"users", ownerId, "quizzes", quizId, "questions", question.Id, "answers", answer.Id}, "/")).
			Set(context.Background(), answer)

		if err4 != nil {
			return err4
		}
	}

	return nil
}
