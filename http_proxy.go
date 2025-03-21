package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers#hop-by-hop_headers
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

var transport = http.Transport{
	MaxIdleConns:          100,
	ExpectContinueTimeout: 1 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	IdleConnTimeout:       90 * time.Second,
}

func proxyDirect(w http.ResponseWriter, r *http.Request) {
	log.Printf("http request from %v to %v", r.RemoteAddr, r.Host)
	pReq, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		log.Printf("error: creating request failed: %v\n", err)
		http.Error(w, "error creating proxy request", http.StatusInternalServerError)
		return
	}

	deleteHopHeaders(r.Header)
	copyHeaders(pReq.Header, r.Header)

	resp, err := transport.RoundTrip(pReq)
	if err != nil {
		log.Printf("error: sending request failed: %v\n", err)
		http.Error(w, "error while sending proxy request", http.StatusServiceUnavailable)
		return
	}

	deleteHopHeaders(resp.Header)
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("error: sending response failed: %v\n", err)
	}
}

func copyHeaders(dst, src http.Header) {
	for name, values := range src {
		for _, value := range values {
			dst.Add(name, value)
		}
	}
}

func deleteHopHeaders(h http.Header) {
	for _, hop := range hopHeaders {
		h.Del(hop)
	}
}
