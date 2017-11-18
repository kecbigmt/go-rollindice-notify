package main

import (
	"fmt"
	"os"
  "net/http"
  "log"

	"github.com/kecbigmt/go-rollindice-notify/workers"
)

func main() {

	go workers.DiscordEventListener()

  http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Hello")
  })

  if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
