package impl

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis"
	goredislib "github.com/redis/go-redis/v9"
	"museum/config"
	"museum/domain"
)

type RedisStateConnector struct {
	RedisClient *goredislib.Client
	RedisPool   redis.Pool
	RedisSync   *redsync.Redsync
	RedisMu     *redsync.Mutex
	Config      config.Config
}

func (rs *RedisStateConnector) WithLock(f func() error) (err error) {
	err = rs.RedisMu.Lock()
	if err != nil {
		return errors.New("could not acquire lock")
	}

	defer (func() {
		_, err = rs.RedisMu.Unlock()
		if err != nil {
			err = errors.New("could not release lock")
		}
	})()

	return f()
}

func (rs *RedisStateConnector) GetExhibits() ([]domain.Exhibit, error) {
	iter := rs.RedisClient.Scan(context.Background(), 0, rs.Config.GetRedisBaseKey()+":exhibit:*", 0).Iterator()
	state := make([]domain.Exhibit, 0)
	for iter.Next(context.Background()) {
		app := domain.Exhibit{}
		key := iter.Val()
		app.Id = key

		res := rs.RedisClient.Get(context.Background(), key)
		if res.Err() == goredislib.Nil {
			return nil, res.Err()
		}

		err := json.Unmarshal([]byte(res.Val()), &app)
		if err != nil {
			return nil, err
		}

		state = append(state, app)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return state, nil
}

func (rs *RedisStateConnector) DeleteExhibitById(id string) error {
	//TODO implement me
	panic("implement me")
}

func (rs *RedisStateConnector) AddExhibit(app domain.Exhibit) error {
	// check if app already exists, if so, return error
	res := rs.RedisClient.Get(context.Background(), rs.Config.GetRedisBaseKey()+":exhibit:"+app.Id)
	if res.Err() == nil {
		return errors.New("exhibit already exists")
	}

	b, err := json.Marshal(app)
	if err != nil {
		return err
	}

	set := rs.RedisClient.Set(context.Background(), rs.Config.GetRedisBaseKey()+":exhibit:"+app.Id, b, 0)
	if set.Err() != nil {
		return set.Err()
	}

	return nil
}
