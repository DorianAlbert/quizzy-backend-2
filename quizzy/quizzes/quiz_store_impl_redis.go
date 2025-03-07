package quizzes

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type RedisCodeResolver struct {
	client *redis.Client
}

func (re *RedisCodeResolver) BindCode(ownerId string, quiz Quiz) error {
	return re.client.Set(context.Background(), quiz.Code, fmt.Sprintf("%s@%s", ownerId, quiz.Id), 0).Err()
}

func (re *RedisCodeResolver) UnbindCode(code string) error {
	return re.client.Del(context.Background(), code).Err()
}

func (re *RedisCodeResolver) GetQuiz(code string) (string, error) {
	return re.client.Get(context.Background(), code).Result()
}

func (re *RedisCodeResolver) IncrRoomPeople(roomId string) error {
	key := fmt.Sprintf("room:%s", roomId)
	if err := re.client.Incr(context.Background(), key).Err(); errors.Is(err, redis.Nil) {
		return re.client.Set(context.Background(), key, 1, 0).Err()
	} else {
		return err
	}
}

func (re *RedisCodeResolver) GetRoomPeople(roomId string) (int, error) {
	return re.client.Get(context.Background(), fmt.Sprintf("room:%s", roomId)).Int()
}

func (re *RedisCodeResolver) ResetRoomPeople(roomId string) error {
	return re.client.Del(context.Background(), fmt.Sprintf("room:%s", roomId)).Err()
}
