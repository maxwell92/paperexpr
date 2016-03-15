package main

import (
    "fmt"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "log"
    "time"
   // "sync/atomic"
    "github.com/garyburd/redigo/redis"
)

//var delay int
//var delay string

var reqQ chan *http.Request
var resQ chan http.ResponseWriter

type RedisConn struct {
    dbid string
    redis.Conn
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprintf(w, "Welcome!\n")
}

func Healthz(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprintf(w, "ok\n")
}

func connectRedis() {
    conn, err := redisConn("127.0.0.1:6379", "root", "0")
    if err != nil {
       log.Fatal("Error: ", err) 
    } 	
   
    delay, err := redis.Int(conn.Do("INCR", "a"))
    if err != nil {
	log.Fatal("Error: ", err)
    }
    
    log.Print("Number: ", delay)
/*    delay, err = redis.String(conn.Do("GET", "a"))
    if err != nil {
	log.Fatal("Error: ", err)
    } 
*/
    conn.FlushClose() 
}

func (r *RedisConn) FlushClose() error {
/*    if r.dbid != "" {
	if _, err := r.Conn.Do("SELECT", r.dbid); err != nil {
	    return nil
	}	
    } 

    if _, err := r.Conn.Do("FLUSHDB"); err != nil {
	return err
    }
*/
    return r.Conn.Close()    
}

func redisConn(host, password, db string) (*RedisConn, error) {
    if host == "" {
	host = ":6379"
    }

    conn, err := redis.DialTimeout("tcp", host, 0, 1 * time.Second, 1 * time.Second)
    if err != nil {
        return nil, err
    }

    if password != "" {
	if _, err := conn.Do("AUTH", password); err != nil {
	    conn.Close()
	    return nil, err
	}
    } 

    if db != "" {
        if _, err := conn.Do("SELECT", db); err != nil {
            conn.Close()
            return nil, err
        }
    }
    
    return &RedisConn{dbid:db, Conn:conn}, nil
}

func Next(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    connectRedis() 
    Send(resQ, reqQ, w, r) 
}

func Send(resQ chan http.ResponseWriter, reqQ chan *http.Request, w http.ResponseWriter, r *http.Request) {
    select {
	case resQ<- w:
    	case reqQ<- r:
    } 
}

func Wait() {
    for {
	select {
	case <- reqQ:
  		Business()
        	
	}
    }
} 

func Business() {
    select {
	case <-time.After(time.Millisecond * 5):
		//log.Print("Business: ", delay);	
    }
}

func main() {
    done := make(chan bool)
    reqQ = make(chan *http.Request)
    resQ = make(chan http.ResponseWriter) 
    router := httprouter.New()
    router.GET("/", Index)
    router.GET("/healthz", Healthz)
    router.GET("/next", Next)
    go Wait()
    log.Fatal(http.ListenAndServe(":9999", router))
    <- done
}
