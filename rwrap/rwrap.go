package rwrap

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	//"strings"
	"time"
)

func newPool(server, port string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server+":"+port)
			if err != nil {
				return nil, err
			}
			/*
			   if _, err := c.Do("AUTH", password); err != nil {
			       c.Close()
			       return nil, err
			   }*/
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

var (
	pool         *redis.Pool
	Port, Server string
	Expire       int
)

func Start() {
	pool = newPool(Server, Port)
}

func typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func Exists(name string) bool {
	conn := pool.Get()
	defer conn.Close()
	exists, _ := redis.Bool(conn.Do("EXISTS", name))
	return exists
}

func HExists(name, field string) bool {
	conn := pool.Get()
	defer conn.Close()
	exists, _ := redis.Bool(conn.Do("HEXISTS", name, field))
	return exists
}

// set boolean
func SetB(name, value bool) {
	conn := pool.Get()
	defer conn.Close()
	conn.Do("SET", name, value)
}

// set hash
func SetH(name, hash, value string) bool {
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("HSET", name, hash, value)
	if err != nil {
		log.Println("Unable to set hash variable")
		log.Println(err)
		return false
	} else {
		if Expire > 0 {
			conn.Do("EXPIRE", name, Expire)
		}
		return true
	}
}

// get hash
func GetH(name, hash string) string {
	conn := pool.Get()
	defer conn.Close()

	s, err := redis.String(conn.Do("HGET", name, hash))

	if err != nil {
		log.Println("Unable to get hash variable")
	}
	return s
}

// mget hash
func MGetH(name string) map[string]string {
	conn := pool.Get()
	defer conn.Close()

	values, err := redis.StringMap(conn.Do("HGETALL", name))

	if err != nil {
		log.Println("Unable to get variable")
	}

	return values
}

// Set multiple hash fields to multiple values
func HMSet(key string, values map[string]string) bool {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(values)...)

	if err != nil {
		log.Println("Unable to HMSET variables")
		log.Println(err)
	}

	if Expire > 0 {
		conn.Do("EXPIRE", key, Expire)
	}

	return true
}

func MGet(keys ...string) []string {
	conn := pool.Get()
	defer conn.Close()

	values, err := redis.Strings(conn.Do("MGET", redis.Args{}.AddFlat(keys)...))

	if err != nil {
		log.Println("Unable to get variable")
	}

	log.Println("%v", values)

	return values
}

// set string
func SetS(name, value string) bool {
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", name, value)
	if err != nil {
		log.Println("Unable to set variable")
		log.Println(err)
		return false
	} else {
		if Expire > 0 {
			conn.Do("EXPIRE", name, Expire)
		}

		return true
	}
}

// del
func Del(name string) bool {
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", name)
	if err != nil {
		log.Println("Unable to remove key " + name)
		log.Println(err)
		return false
	} else {
		log.Println("Removed ", name)
		return true
	}
}

// get string
func GetS(name string) string {
	conn := pool.Get()
	defer conn.Close()
	s, err := redis.String(conn.Do("GET", name))

	if err != nil {
		log.Println("Unable to get variable")
	}
	return s
}

func HINCRBY(hash, field string, value int) bool {
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("HINCRBY", hash, field, value)
	if err != nil {
		log.Println("Unable to HINCRBY variable")
		log.Println(err)
		return false
	} else {
		return true
	}
}

func SADD(set, value string) bool {
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("SADD", set, value)
	if err != nil {
		log.Println("Unable to SADD variable")
		log.Println(err)
		return false
	} else {
		if Expire > 0 {
			conn.Do("EXPIRE", set, Expire)
		}

		return true
	}
}
