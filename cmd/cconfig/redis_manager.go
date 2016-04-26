package main

import (
	"github.com/garyburd/redigo/redis"
	"github.com/juju/errors"
	"github.com/CodisLabs/codis/pkg/utils"
	log "github.com/ngaut/logging"
	"strings"
	"sync"
)

// del data
func getData(network string, valKey string) (string, error) {
	conn, err := utils.DialTo(network, globalEnv.Password())
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
func delData(network string, valKey string) (int, error) {
	conn, err := utils.DialTo(network, globalEnv.Password())
	defer conn.Close()
	if (err != nil) {
		log.Warning(err)
	}

	if !strings.HasSuffix(valKey, "*") && !strings.HasPrefix(valKey, "*") {
		// del data
		_, err := conn.Do("DEL", valKey)
		if err != nil {
			return -1, errors.Trace(err)
		}
	} else {
		scanAndDel(conn, valKey, 0)
	}

	return 0, nil
}

// flush dba
func flushDB(network string) (int, error)  {
	conn, err := utils.DialTo(network, globalEnv.Password())
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
	conn, err := utils.DialTo(network, globalEnv.Password())
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

// set data
func stopRedis(network string) (int, error) {
	conn, err := utils.DialTo(network, globalEnv.Password())
	defer conn.Close()
	if (err != nil) {
		log.Warning(err)
	}
	_, _ = conn.Do("SHUTDOWN")

	return 0, nil
}

func scanAndDel(conn redis.Conn, key string, cursor int) {
	result, err := redis.Values(conn.Do("SCAN", cursor, "MATCH", key, "COUNT", "1000"))
	if (err != nil) {
		log.Error(err)
	}
	var wg sync.WaitGroup

	if result != nil {
		cursor, _ = redis.Int(result[0], nil);
		values, _ := redis.Strings(result[1], nil);

		wg.Add(1)

		go func(conn redis.Conn, values []string) {
			defer wg.Done()
			for _, k := range values {
				_, err := conn.Do("DEL", k)
				if (err != nil) {
					log.Error(err)
				}
			}

		}(conn, values)

		if cursor != 0 {
			scanAndDel(conn, key, cursor)
		}
	}

	wg.Wait()
}