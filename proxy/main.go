package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"io"
	"kadam.net/test_task"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	port          = flag.Int("p", 0, "app port")
	rawRecipients = flag.String("r", "", "comma-separated list of recipient ports, e.g. 8051,8052,8053")
)

const (
	reqTimeout = time.Duration(100) * time.Millisecond

	readTimeout         = time.Duration(500) * time.Millisecond
	writeTimeout        = time.Duration(500) * time.Millisecond
	maxIdleConnDuration = time.Duration(1) * time.Hour
)

func init() {
	flag.Parse()
	if *port == 0 {
		log.Fatal("Invalid port", *port)
	}

	if *rawRecipients == "" {
		log.Fatal("Invalid recipients", *rawRecipients)
	}
}

func main() {
	log.Println("Initializing proxy...")
	proxy := Proxy{}

	proxy.init(*rawRecipients, proxyInitParams{
		readTimeout:         readTimeout,
		writeTimeout:        writeTimeout,
		maxIdleConnDuration: maxIdleConnDuration,
	})

	log.Println("Starting proxy at port", *port)
	proxy.start()
}

type Proxy struct {
	recipients []string
	httpClient *fasthttp.Client
}

func (p *Proxy) init(rawRecipients string, params proxyInitParams) {
	p.recipients = parseRecipients(rawRecipients)

	p.httpClient = &fasthttp.Client{
		ReadTimeout:                   params.readTimeout,
		WriteTimeout:                  params.writeTimeout,
		MaxIdleConnDuration:           params.maxIdleConnDuration,
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}
}

func (p *Proxy) start() {
	// Not using router because there's just one endpoint for a server.
	// but router usage (such as gorilla/mux or fasthttp-routing) is a good practice.
	log.Fatal(fasthttp.ListenAndServe(":"+strconv.Itoa(*port), p.fastHttpProxyHandler))
}

func (p *Proxy) fastHttpProxyHandler(ctx *fasthttp.RequestCtx) {
	var (
		commonRequest test_task.CommonProxyRequest
	)
	err := json.Unmarshal(ctx.PostBody(), &commonRequest)
	if err != nil {
		log.Println("unmarshal common request err: ", err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	bestBidResponse, err := p.sendToRecipients(commonRequest)
	if err != nil {
		log.Println("send to recipients err: ", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(bestBidResponse)
	if err != nil {
		log.Println("marshal best bid response err: ", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetBody(b)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// proxyHandler is not used, but I wanted to highlight net/http as common approach
func (p *Proxy) proxyHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("read request body err: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var (
		commonRequest test_task.CommonProxyRequest
	)
	err = json.Unmarshal(body, &commonRequest)
	if err != nil {
		log.Println("unmarshal common request err: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bestBidResponse, err := p.sendToRecipients(commonRequest)
	if err != nil {
		log.Println("send to recipients err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(bestBidResponse)
	if err != nil {
		log.Println("marshal best bid response err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, string(b))
}

func (p *Proxy) sendToRecipients(commonRequest test_task.CommonProxyRequest) (*test_task.InnerResponse, error) {
	var (
		wg sync.WaitGroup
	)

	commonReqBytes, err := json.Marshal(commonRequest)
	if err != nil {
		log.Println("unable to marshal common request", err)
		return nil, err
	}

	innerReq := commonToInnerRequest(commonRequest)

	innerReqBytes, err := json.Marshal(innerReq)
	if err != nil {
		log.Println("unable to marshal inner request", err)
		return nil, err
	}

	recipientResponses := make([]test_task.InnerResponse, 0)

	wg.Add(len(p.recipients))
	for _, recipient := range p.recipients {
		go func(recipient string) {
			defer wg.Done()
			req := fasthttp.AcquireRequest()
			req.SetRequestURI(recipient)
			req.Header.SetMethod(fasthttp.MethodPost)
			req.Header.SetContentTypeBytes(headerContentTypeJson)
			req.SetBodyRaw(innerReqBytes)

			resp := fasthttp.AcquireResponse()
			err := p.httpClient.DoTimeout(req, resp, reqTimeout)
			fasthttp.ReleaseRequest(req)
			defer fasthttp.ReleaseResponse(resp)
			if err != nil {
				log.Println("req:", string(innerReqBytes), "err", err)

				return
			}

			respBody := resp.Body()

			var innerResp test_task.InnerResponse
			err = json.Unmarshal(respBody, &innerResp)
			if err != nil {
				log.Println("unable to unmarshal inner response, err:", err, "body", string(respBody))
				return
			}

			recipientResponses = append(recipientResponses, innerResp)
		}(recipient)
	}
	wg.Wait()

	var (
		maxIndex int
		maxValue float64

		maxBidResponse []byte
	)
	for i, recipientResponse := range recipientResponses {
		if i == 0 || recipientResponse.Bid > maxValue {
			maxIndex = i
			maxValue = recipientResponse.Bid
		}
	}

	recipientResponsesBytes, err := json.Marshal(recipientResponses)
	if err != nil {
		log.Println("unable to marshal recipient responses", err)
		return nil, err
	}

	if len(recipientResponses) != 0 {
		maxBidResponse, err = json.Marshal(recipientResponses[maxIndex])
		if err != nil {
			log.Println("unable to marshal max bid response", err)
			return nil, err
		}
	}

	log.Printf("req: %s; inner: %s;\nrecip: %s\nmax: %s\n", string(commonReqBytes), string(innerReqBytes),
		string(recipientResponsesBytes), string(maxBidResponse))

	return &recipientResponses[maxIndex], nil
}
