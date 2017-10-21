package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Response struct {
	Arm string `json:"arm"`
}

type UpdateRequest struct {
	Arm    int `json:"arm"`
	Reward int `json:"reward"`
}

func selectArm(url string) (*Response, error) {
	var body Response
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, err
	}
	return &body, nil
}

func update(url string, arm, reward int) error {
	client := &http.Client{}

	data, err := json.Marshal(UpdateRequest{
		Arm:    arm,
		Reward: reward,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Println("got response:", string(b))
	return nil
}

func experiment() error {
	id := "4d910d09-7b3a-4084-8493-5bc67904c8f8"
	url := fmt.Sprintf("http://localhost:3000/v1/experiments/%s/arms", id)

	resp, err := selectArm(url)
	if err != nil {
		return err
	}
	arm := 0

	if resp.Arm == "" {
		arm = 0
	} else {
		i, err := strconv.Atoi(resp.Arm)
		if err != nil {
			return err
		}
		arm = i
	}
	reward := 0
	if rand.Float64() > 0.5 {
		reward = 1
	}

	err = update(url, arm, reward)
	if err != nil {
		return err
	}
	return nil
}
func main() {
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup
	n := 1000
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			err := experiment()
			if err != nil {
				log.Println(err)
			}
		}()
	}
	wg.Wait()
	log.Println("done")
}
