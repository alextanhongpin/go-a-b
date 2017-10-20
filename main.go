package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/alextanhongpin/go-a-b/bandit"

	"github.com/julienschmidt/httprouter"
)

var experiments *bandit.Experiment

func main() {
	rand.Seed(time.Now().UnixNano())
	experiments = bandit.NewExperiment()

	log.Printf("got %#v", exp)

	r := httprouter.New()
	r.GET("/", index)
	r.GET("/experiments", getExperiments)
	r.POST("/experiments", createExperiment)
	r.PATCH("/experiments/:id", updateExperiment)
	r.GET("/arms/:id", getArm)
	r.PATCH("/arms/:id", updateArm)

	log.Println("listening to port *:8080. press ctrl + c to cancel")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, `{"ok": true}`)
}

type GetAllResponse struct {
	Data []bandit.Response `json:"data"`
}

func getExperiments(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	res := GetAllResponse{
		Data: experiments.All(),
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
func getArm(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")

	exp := experiments.One(id)
	arm, exploit := exp.SelectArm()

	res := make(map[string]interface{})
	res["arm"] = arm
	res["exploit"] = exploit

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func updateArm(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	arm := ps.ByName("arm")
	reward := ps.ByName("reward")

	armInt, _ := strconv.Atoi(arm)
	rewardInt, _ := strconv.Atoi(reward)

	exp := experiments.One(id)
	exp.Update(armInt, float64(rewardInt))

	res := make(map[string]interface{})
	res["arm"] = armInt
	res["reward"] = rewardInt
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func updateExperiment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	exp := experiments.One(id)
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	for i := 0; i < 1000; i++ {
		var reward int
		if r1.Float64() > 0.5 {
			reward = 1
		}
		arm, _ := exp.SelectArm()
		exp.Update(arm, float64(reward))
	}
	res := make(map[string]interface{})
	res["ok"] = true
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func createExperiment(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	exp := experiments.NewEpsilonGreedy(3, 0.1)
	if err := json.NewEncoder(w).Encode(exp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
