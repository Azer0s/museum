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

func (rs *RedisStateConnector) withLock(f func()) (err error) {
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

	f()
	return nil
}

func (rs *RedisStateConnector) GetApplications() (state []domain.Application, err error) {
	lockErr := rs.withLock(func() {
		iter := rs.RedisClient.Scan(context.Background(), 0, rs.Config.GetRedisBaseKey()+":state:app:*", 0).Iterator()

		for iter.Next(context.Background()) {
			app := domain.Application{}
			key := iter.Val()
			app.Id = key

			res := rs.RedisClient.Get(context.Background(), key)
			if res.Err() == goredislib.Nil {
				state = nil
				err = res.Err()
				return
			}

			err := json.Unmarshal([]byte(res.Val()), &app)
			if err != nil {
				state = nil
				return
			}

			state = append(state, app)
		}

		if err := iter.Err(); err != nil {
			state = nil
			return
		}

		err = nil
		return
	})

	if lockErr != nil {
		return nil, lockErr
	}
	return
}

func (rs *RedisStateConnector) DeleteApplication(app domain.Application) error {
	//TODO implement me
	panic("implement me")
}

func (rs *RedisStateConnector) AddApplication(app domain.Application) error {
	//TODO implement me
	panic("implement me")
}
