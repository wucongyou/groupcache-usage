package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/golang/groupcache"
)

func main() {
	fmt.Println("ok")
	self := (&url.URL{Scheme: "http", Host: ":8081"}).String()
	pool := groupcache.NewHTTPPool(self)
	peers := [...]*url.URL{&url.URL{Scheme: "http", Host: ":8082"}, &url.URL{Scheme: "http", Host: ":8083"}}
	peerStrs := make([]string, 0, len(peers))
	for _, peer := range peers {
		peerStrs = append(peerStrs, peer.String())
	}
	pool.Set(peerStrs...)
	gp := groupcache.NewGroup("default", 64<<20, groupcache.GetterFunc(func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
		time.Sleep(time.Second * 3)
		dest.SetString("result for: " + key)
		return nil
	}))
	fmt.Println(gp)
	http.HandleFunc("/default", func(rw http.ResponseWriter, r *http.Request) {
		k := r.URL.Query().Get("key")
		fmt.Printf("user get %s from groupcache\n", k)
		var data []byte
		gp.Get(nil, k, groupcache.AllocatingByteSliceSink(&data))
		rw.Write(data)
	})
	http.ListenAndServe(":8081", nil)
}
