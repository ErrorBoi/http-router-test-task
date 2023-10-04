package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"

	"kadam.net/test_task"
)

var (
	proxy = flag.Int("p", 8050, "Proxy app port.")
	limit = flag.Int("l", 1e6, "Request count limit.")
)

func init() {
	flag.Parse()
	if *proxy == 0 {
		log.Fatal("Wrong proxy app port ", *proxy)
	}
	if *limit == 0 {
		log.Fatal("Wrong request limit ", *limit)
	}
}

func main() {
	var (
		wg   sync.WaitGroup
		addr = "http://localhost:" + strconv.Itoa(*proxy) + test_task.ProxyEndpoint
	)
	netClient := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost: 50,
		},
	}
	log.Println("Start client session on", addr, "limit", *limit)
	for i := 0; i < *limit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := test_task.GenerateRandomRequest(rand.Intn(4))
			resp, err := netClient.Post(addr, "application/json", bytes.NewBuffer(req))
			if err != nil {
				log.Println("req:", string(req), "err", err)
				return
			}
			body, _ := io.ReadAll(resp.Body)
			log.Println("req:", string(req), "resp", string(body))
		}()
	}
	wg.Wait()
}
