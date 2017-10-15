package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis"
)

const port = ":8080"

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:21586",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(pong)

	mux := http.NewServeMux()

	ab := make(map[string]interface{})
	ab["name"] = "hello-test"
	ab["option-1"] = "yes"
	ab["option-2"] = "no"
	client.HMSet("test-ab", ab)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	})

	log.Printf("listening to port *%v. press ctrl + c to cancel.\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
