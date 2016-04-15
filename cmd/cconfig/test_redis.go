package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/wandoulabs/codis/pkg/utils/errors"
	"sync"
	"strings"
	"encoding/base64"
	"hash/crc32"
)

func main() {
	//conn, err := utils.DialTo("132.35.224.78:6381", "")
	//
	//defer conn.Close()
	//if (err != nil) {
	//	fmt.Println(err)
	//}
	//
	//for i := 0; i < 500; i++ {
	//	conn.Do("SET", "test" + strconv.Itoa(i), "1234dd"+ strconv.Itoa(i))
	//}


	users := strings.Split("admin@123, yangdaix@123", ",")
	fmt.Println(users[0])
	for _, v := range users {
		fmt.Println(strings.TrimSpace(v))
	}

	fmt.Println(base64.StdEncoding.EncodeToString([]byte("API:KEY")))

	//var yesNo string
	//fmt.Println("DDDD")
	//fmt.Scanln(&yesNo)
	//
	//fmt.Println(yesNo)

	fmt.Println(int(crc32.ChecksumIEEE([]byte{1,2,3}) % 1024))
	fmt.Println(2048%1024)
	slot :=0
	slot ++
	fmt.Println(slot)

	//arr, err := redis.Values(conn.Do("SCAN", 0, "MATCH", "test1*"))
	//if (err != nil) {
	//	fmt.Println(err)
	//}
	//
	//
	//fmt.Println(redis.String(arr[0], nil))
	//dd,_ := redis.Strings(arr[1], nil);
	//fmt.Println(dd)
	//for _,a := range dd {
	//	fmt.Println(redis.String(a, nil))
	//}
	//fmt.Println(arr)

	  //var varlues = make([]string, 0)
	//values := scan(conn, "t*", 0)
	//
	//
	//
	//fmt.Println(values)



	//ret, err := scanAndDel1(conn, "t*", 0)
	//if (err != nil) {
	//	fmt.Println(err)
	//}



	//fmt.Println(ret)

}

func scan(conn redis.Conn, key string, cursor int) (v []string) {
	result, err := redis.Values(conn.Do("SCAN", cursor, "MATCH", key, "COUNT", "1000"))
	if (err != nil) {
		fmt.Println(err)
	}

	var retArray = make([]string, 0);

	if result != nil {
		cursor, _ = redis.Int(result[0], nil);
		values, _ := redis.Strings(result[1], nil);
		for _, v := range values {
			retArray = append(retArray, v)
		}
		if cursor != 0 {
			temps := scan(conn, key, cursor)
			for _, v1 := range temps {
				retArray = append(retArray, v1)
			}

		}
	}



	return retArray
}

func scanAndDel1(conn redis.Conn, key string, cursor int) (int, error) {
	result, err := redis.Values(conn.Do("SCAN", cursor, "MATCH", key, "COUNT", "1000"))
	if (err != nil) {
		return -1, errors.Trace(err)
	}
	var wg  sync.WaitGroup
	if result != nil {
		cursor, _ = redis.Int(result[0], nil);
		values, _ := redis.Strings(result[1], nil);
		wg.Add(1)
		 go func(conn redis.Conn, values []string) (int, error) {
			 defer wg.Done()
			for _, key := range values {
				fmt.Println(key + "ddddddddddd")
				_, err := conn.Do("DEL", key)
				if (err != nil) {
					return -1, errors.Trace(err)
				}
			}

			return 0, nil
		}(conn, values)

		if cursor != 0 {
			return scanAndDel1(conn, key, cursor)
		}
	}

	wg.Wait()

	return 0, nil
}
