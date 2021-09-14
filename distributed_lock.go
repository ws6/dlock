package dlock

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
)

type DMutex struct {
	mutex *redsync.Mutex

	*Dlock
	onLockSuccess func()
	onUnlock      func()
}

type Dlock struct {
	config map[string]string
	// configStr like redis server addr,pool size,password,dbnum,IdleTimeout second
	// e.g. 127.0.0.1:6379,100,password,0,30
	// or 127.0.0.1:6379 without password
	configStr string
	rs        *redsync.Redsync
	Close     func() error
}

func NewDlock(cfg map[string]string) (*Dlock, error) {
	ret := new(Dlock)
	ret.Close = func() error {
		fmt.Println(`not initialized yet`)
		return nil
	}
	ret.config = make(map[string]string)

	// address
	//copy the config
	for k, v := range cfg {
		ret.config[k] = v
	}

	//prepare connecting to redis
	ret.configStr = ret.config[`redis_config_string`]

	rpool, err := getRedisPool(ret.configStr)
	if err != nil {
		return nil, fmt.Errorf(`getRedisPool:%s`, err.Error())
	}
	ret.Close = rpool.Close
	rs := redigo.NewPool(rpool)
	ret.rs = redsync.New(rs)

	return ret, nil
}

func (self *Dlock) GetExpireSecond() int {
	expirySecond := 15
	if n, err := strconv.Atoi(self.config[`expire_second`]); err == nil && n > 0 {
		expirySecond = n
	}
	return expirySecond
}

func (self *Dlock) GetNumRetryOnAcquiringLock() int {
	numRetry := 10
	if n, err := strconv.Atoi(self.config[`num_retry`]); err == nil && n > 0 {
		numRetry = n
	}

	return numRetry
}

func (self *Dlock) NewMutex(ctx context.Context, key string) *DMutex {
	ret := new(DMutex)
	ret.Dlock = self

	ret.mutex = self.rs.NewMutex(key,
		redsync.WithExpiry(time.Duration(self.GetExpireSecond())*time.Second),
		redsync.WithTries(self.GetNumRetryOnAcquiringLock()),
	)

	//renew automatically until explicitly canceled or unlock
	ret.onLockSuccess = func() {
		_ctx, canclFn := context.WithCancel(ctx)
		ret.onUnlock = func() {
			canclFn()

		}
		renewInterval := self.GetExpireSecond() / 3
		if renewInterval <= 0 {
			renewInterval = 1
		}
		for {

			select {
			case <-_ctx.Done():

				return
			case <-time.After(time.Second * time.Duration(renewInterval)):
				b, err := ret.mutex.Extend()
				if !b || err != nil {
					fmt.Println(`Extend err`, renewInterval, `seconds`, b, err)
				}

			}
		}

	}

	return ret
}

func (self *DMutex) Lock() error {

	ret := self.mutex.Lock()
	if ret != nil {
		return ret
	}
	if self.onLockSuccess != nil {
		go self.onLockSuccess()
	}

	return nil
}

func (self *DMutex) Unlock() error {

	if self.onUnlock != nil {
		self.onUnlock()
	}

	ok, err := self.mutex.Unlock()
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf(`unlock:not all nodes returned success`)
	}

	return nil
}
