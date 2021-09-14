package dlock

import (
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

//getRedisPool return a redis pool
// configStr like redis server addr,pool size,password,dbnum,IdleTimeout second
// e.g. 127.0.0.1:6379,100,password,0,30
// or 127.0.0.1:6379 without password

func getRedisPool(configStr string) (*redis.Pool, error) {
	connStr := configStr
	configs := strings.Split(configStr, ",")
	dbNum := 0
	idleTimeoutSecond := 0
	poolSize := 100
	password := "" //no password
	if len(configs) > 0 {
		connStr = configs[0]
	}
	if len(configs) > 1 {
		if n, err := strconv.Atoi(configs[1]); err == nil && n > 0 {
			poolSize = n
		}
	}
	if len(configs) > 2 {
		password = configs[2]
	}
	if len(configs) > 3 {
		if n, err := strconv.Atoi(configs[3]); err == nil && n > 0 {
			dbNum = n
		}
	}

	if len(configs) > 4 {
		if n, err := strconv.Atoi(configs[4]); err == nil && n > 0 {
			idleTimeoutSecond = n
		}
	}

	ret := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", connStr)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err = c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			// some redis proxy such as twemproxy is not support select command
			if dbNum > 0 {
				_, err = c.Do("SELECT", dbNum)
				if err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		MaxIdle: poolSize,
	}
	ret.IdleTimeout = time.Duration(idleTimeoutSecond) * time.Second

	return ret, nil
}
