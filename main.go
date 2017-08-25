package main

import (
	"encoding/json"
	"fmt"
	"net/http"

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
		fmt.Printf("user get %s from groupcache\n", k)
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
	http.ListenAndServe(":8081", nil)
}
