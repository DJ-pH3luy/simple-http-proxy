package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"sync"
)

func proxyConnect(w http.ResponseWriter, r *http.Request) {
	log.Printf("connect request from %v to %v", r.RemoteAddr, r.Host)
	target, err := net.Dial("tcp", r.Host)
	if err != nil {
		log.Printf("error: dialing %v failed: %v", r.Host, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Fatal("server error - could not create hijacker")
	}

	client, _, err := hj.Hijack()
	if err != nil {
		log.Fatalf("server error - hijacking failed: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go tunnel(&wg, client, target)
	go tunnel(&wg, target, client)
	
	wg.Wait()
	client.Close()
	target.Close()
	log.Println("closed connection to", r.Host)
}

func tunnel(wg *sync.WaitGroup, dst io.WriteCloser, src io.ReadCloser) {
	defer wg.Done()

	_, err := io.Copy(dst, src)
	if err != nil {
		log.Printf("error while tunneling: %v", err)
	}
}
