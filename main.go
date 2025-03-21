package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	var address = flag.String("a", "0.0.0.0:8080", "proxy server address")
	flag.Parse()

	log.Printf("start listening on %v\n", *address)
	if err := http.ListenAndServe(*address, &proxyHandler{}); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
}

type proxyHandler struct {
}

func (p *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		proxyConnect(w, r)
	} else {
		proxyDirect(w, r)
	}
}
