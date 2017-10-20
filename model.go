package main

type Meta map[string]interface{}

type CreateExperimentRequest struct {
	Type string `json:"type"` // The algorithm used
	// N       int     `json:"n"` // The number of arms
	Epsilon     float64   `json:"epsilon"`     // The epsilon value
	Description string    `json:"description"` // The description of the experiment
	Name        string    `json:"name"`        // The name of the experiment
	Features    []string  `json:"features"`    // The features to be tested
	Reward      float64   `json:"reward"`      // The value to increment if the user perform certain action
	Counts      []int64   `json:"counts"`      // Optional. The starting counts loaded from elsewhere
	Rewards     []float64 `json:"rewards"`     // Optional. The starting rewards loaded from elsewhere
	Meta        Meta      `json:"meta"`        // Additional meta information
}
