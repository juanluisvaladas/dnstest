package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func worker(host string, quit <-chan int, selfKill chan<- int) {
	for {
		time.Sleep(200 * time.Millisecond)
		select {
		case <-quit:
			return
		default:
			addr, err := net.LookupHost(host)
			wgDone := false

			if err != nil {
				log.Println(err)
				wgDone = true
			}
			if len(addr) == 0 {
				log.Println("addr len = 0")
				wgDone = true
			}
			containsARecord := false
			for _, a := range addr {
				if strings.ContainsAny(a, ".") {
					containsARecord = true
				}
			}
			if !containsARecord {
				log.Println("Did not get an A record")
				log.Println("addr")
				wgDone = true
			}

			if wgDone {
				selfKill <- 0
			}
		}
	}
}

func main() {

	wn, err := strconv.Atoi(os.Getenv("WORKERS"))
	host := os.Getenv("LOOKUP_ADDR")
	if err != nil {
		log.Fatal("Cannot parse WORKERS: ", err)
	}

	if wn < 1 {
		log.Fatal("WORKERS must be at least 1")
	}

	if len(host) < 1 {
		log.Fatal("LOOKUP_ADDR is empty")
	}

	// once we hit an error we want to stop all the workers
	quit := make(chan int)
	selfKill := make(chan int)

	for i := 0; i < wn; i++ {
		go worker(host, quit, selfKill)
	}
	<-selfKill
	close(quit)

	// And we don't want to exit, we'll leave it sleeping forever,
	// so that if we can set restartPolicy = always and restart the pod
	// By simply running: oc delete pod <pod name>
	log.Println("Terminating goroutines and waiting")
	for {
		time.Sleep(time.Hour)
	}
}
