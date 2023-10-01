package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"kadam.net/test_task"
)

var (
	id   = flag.Int("i", 0, "Recipient ID.")
	port = flag.Int("p", 0, "App port.")
)

func init() {
	flag.Parse()
	if *id == 0 {
		log.Fatal("Wrong recipient ID ", *id)
	}
	if *port == 0 {
		log.Fatal("Wrong port ", *port)
	}
}

func main() {
	addr := ":" + strconv.Itoa(*port)
	http.HandleFunc(test_task.BidEndpoint, handler)
	log.Printf("Starting recipient #%d at %s", *id, addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("err: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var (
		req  test_task.InnerRequest
		resp test_task.InnerResponse
	)
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Println("err: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp.Id = req.Id
	resp.Message = test_task.GetRandomStr(100)
	resp.RecipId = int32(*id)
	resp.Bid = test_task.GetRandomFloat(req.MinPrice)
	b, err := json.Marshal(resp)
	if err != nil {
		log.Println("err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("req:", string(body), "; res:", string(b))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, string(b))
}
