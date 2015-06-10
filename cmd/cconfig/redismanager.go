package main

import (
	"github.com/garyburd/redigo/redis"
	"github.com/juju/errors"
	log "github.com/ngaut/logging"
)

// del data
func getData(network string, valKey string) (string, error) {
	conn, err := redis.Dial("tcp", network)
	defer conn.Close()
	if (err != nil) {
		log.Warning(err)
	}
	// del data
	data,err := redis.String(conn.Do("GET", valKey))
	if err != nil {
		return "", errors.Trace(err)
	}

	return data, nil
}

// del data
func delData(network string, valKey string) (int, error)  {
	conn, err := redis.Dial("tcp", network)
	defer conn.Close()
	if (err != nil) {
		log.Warning(err)
	}
	// del data
	_, err = conn.Do("DEL", valKey)
	if err != nil {
		return -1, errors.Trace(err)
	}
	return 0, nil
}

// flush dba
func flushDB(network string) (int, error)  {
	conn, err := redis.Dial("tcp", network)
	defer conn.Close()
	if (err != nil) {
		log.Warning(err)
	}
	_, err = conn.Do("FLUSHDB")
	if err != nil {
		return -1, errors.Trace(err)
	}

	return 0, nil
}

// set data
func setData(network string, key string, val string) (int, error) {
	conn, err := redis.Dial("tcp", network)
	defer conn.Close()
	if (err != nil) {
		log.Warning(err)
	}
	_, err = conn.Do("SET", key, val)
	if err != nil {
		return -1, errors.Trace(err)
	}
	return 0, nil
}