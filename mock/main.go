package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// Mock server, при проверке состояния которого он может вернуть с 20% вероятностью not found

func main() {

	name := flag.String("name", "1", "")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from mock backend "+*name)
	})
	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if rand.Intn(100) < 20 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "not found")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	fmt.Println("Mock server listening on :9001")
	http.ListenAndServe(":9001", nil)
}
