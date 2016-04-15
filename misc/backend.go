package misc

import (
	"time"

	log "code.google.com/p/log4go"
	"github.com/garyburd/redigo/redis"
	"github.com/pmylund/go-cache"
)

type backend struct {
	SessStore *cache.Cache
	ClStore   *redis.Pool
}

var Backend *backend

func InitBackend() error {
	Backend = &backend{
		SessStore: cache.New(Conf.Cache.DefaultExpire*time.Hour, Conf.Cache.GCInterval*time.Hour),
		ClStore: &redis.Pool{
			MaxActive:   Conf.Redis.MaxActive,
			MaxIdle:     Conf.Redis.MaxIdle,
			IdleTimeout: Conf.Redis.IdleTimeout * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.DialTimeout(
					"tcp", Conf.Redis.Addr,
					Conf.Redis.ConnectTimeout*time.Second,
					Conf.Redis.ReadTimeout*time.Second,
					Conf.Redis.WriteTimeout*time.Second)
				if err != nil {
					log.Warn("failed to connect Redis, (%s)", err)
					return nil, err
				}
				if Conf.Redis.Passwd != "" {
					if _, err := c.Do("AUTH", Conf.Redis.Passwd); err != nil {
						log.Warn("failed to auth Redis, (%s)", err)
						return nil, err
					}
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		},
	}
	return nil
}
