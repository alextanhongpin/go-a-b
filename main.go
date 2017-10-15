package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
)

const port = ":8080"

type AB struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Options     []string `json:"options"`
	OptionsCSV  string   `json:"options_csv"`
	Scores      []int64  `json:"scores"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
}

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(pong)

	mux := http.NewServeMux()

	// ab := make(map[string]interface{})
	// ab["name"] = "hello-test"
	// ab["option-1"] = "yes"
	// ab["option-2"] = "no"
	// client.HMSet("test-ab", ab)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {

			path := strings.TrimPrefix(r.URL.Path, "/") //strings.Join(strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/"), "-")

			var ab AB
			if err := json.NewDecoder(r.Body).Decode(&ab); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			ab.Slug = path
			ab.CreatedAt = time.Now().UnixNano() / 1000000
			ab.UpdatedAt = time.Now().UnixNano() / 1000000
			ab.ID = uuid.NewV4().String()
			ab.OptionsCSV = strings.ToLower(strings.Join(ab.Options, ","))

			// Convert it back to map[string] interface{}

			in, err := json.Marshal(ab)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			out := make(map[string]interface{})
			if err := json.Unmarshal(in, &out); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			for _, v := range ab.Options {
				out[strings.ToLower(v)] = 0
			}

			delete(out, "options")

			cmd := client.HMSet(path, out)
			if cmd.Err() != nil {
				http.Error(w, cmd.Err().Error(), http.StatusBadRequest)
				return
			}
			val := cmd.Val()
			log.Println(val)

			fmt.Fprintf(w, `{"body": "%v"}`, val)
		} else if r.Method == http.MethodGet {
			path := strings.TrimPrefix(r.URL.Path, "/")

			cmd := client.HGetAll(path)
			if cmd.Err() != nil {
				http.Error(w, cmd.Err().Error(), http.StatusBadRequest)
				return
			}

			val := cmd.Val()

			createdAt, err := strconv.Atoi(val["created_at"])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			updatedAt, err := strconv.Atoi(val["updated_at"])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			options := strings.Split(val["options_csv"], ",")
			scores := make([]int64, len(options))
			for k, opt := range options {
				val, err := strconv.Atoi(val[opt])
				if err != nil {
					scores[k] = 0
				} else {
					scores[k] = int64(val)
				}
			}

			ab := AB{
				Name:        val["name"],
				Description: val["description"],
				Options:     strings.Split(val["options_csv"], ","),
				Scores:      scores,
				CreatedAt:   int64(createdAt),
				UpdatedAt:   int64(updatedAt),
				ID:          val["id"],
				Slug:        val["slug"],
			}

			if err := json.NewEncoder(w).Encode(ab); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			// http.Error(w, http.ErrNotSupported.Error(), http.StatusBadRequest)
		} else if r.Method == http.MethodPut {
			path := strings.TrimPrefix(r.URL.Path, "/")
			cmd := client.HIncrBy(path, "yes", 1)
			if cmd.Err() != nil {
				http.Error(w, cmd.Err().Error(), http.StatusBadRequest)
				return
			}
			fmt.Fprint(w, `{"ok": true}`)
		}
	})

	log.Printf("listening to port *%v. press ctrl + c to cancel.\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
