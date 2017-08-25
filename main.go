package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/groupcache"
)

func main() {
	self := ":8081"
	pool := groupcache.NewHTTPPool(self)
	peers := []string{":8082", "8083"}
	pool.Set(peers...)
	gp := groupcache.NewGroup("default", 64<<20, groupcache.GetterFunc(func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
		dest.SetString("result for: " + key)
		return nil
	}))
	http.HandleFunc("/default", func(rw http.ResponseWriter, r *http.Request) {
		k := r.URL.Query().Get("key")
		var data []byte
		gp.Get(nil, k, groupcache.AllocatingByteSliceSink(&data))
		rw.Write(data)
	})
	http.HandleFunc("/stats", func(rw http.ResponseWriter, r *http.Request) {
		statsStr, err := json.Marshal(gp.Stats)
		if err != nil {
			rw.Write([]byte("server error"))
			return
		}
		rw.Write([]byte(statsStr))
	})
	go http.ListenAndServe(":8081", nil)
	go http.ListenAndServe(":8082", nil)
	go http.ListenAndServe(":8083", nil)
	fmt.Println("start ok")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		s := <-ch
		fmt.Println("got a signal: " + s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			fmt.Println("quit")
			return
		default:
			return
		}
	}
}
