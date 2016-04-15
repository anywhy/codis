package main

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"github.com/juju/errors"
	"github.com/CodisLabs/codis/pkg/utils"
)

type AliveChecker interface {
	CheckAlive() error
}

var (
	_ AliveChecker = &redisChecker{}
)

type redisChecker struct {
	addr           string
	passwd         string
	defaultTimeout time.Duration
}

func (r *redisChecker) ping() error {
	c, err := utils.DialToTimeout(r.addr, r.passwd, r.defaultTimeout, r.defaultTimeout)
	if err != nil {
		return err
	}

	defer c.Close()
	_, err = c.Do("ping")
	return err
}

func (r *redisChecker) CheckAlive() error {
	var err error
	for i := 0; i < 2; i++ { //try a few times
		err = r.ping()
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}

		return nil
	}

	return err
}
