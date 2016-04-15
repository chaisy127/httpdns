package handler

import (
	"fmt"
	"github.com/garyburd/redigo/redis"

	"httpdns/misc"
)

type Cache struct {
}

func (c *Cache) Set(url, host, ip string) error {
	conn := misc.Backend.ClStore.Get()
	if conn == nil {
		return fmt.Errorf("failed to get redis connection")
	}
	defer conn.Close()

	key := fmt.Sprintf("%s_host", url)
	_, err := conn.Do("SETEX", key, misc.Conf.Ttl, host)
	if err != nil {
		return err
	}

	key = fmt.Sprintf("%s_ip", url)
	_, err = conn.Do("SETEX", key, misc.Conf.Ttl, ip)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) Get(url string) (ip, host string, err error) {
	conn := misc.Backend.ClStore.Get()
	if conn == nil {
		return "", "", fmt.Errorf("failed to get redis connection")
	}
	defer conn.Close()

	key := fmt.Sprintf("%s_host", url)
	host, _ = redis.String(conn.Do("GET", key))

	key = fmt.Sprintf("%s_ip", url)
	ip, _ = redis.String(conn.Do("GET", key))

	return ip, host, err
}
