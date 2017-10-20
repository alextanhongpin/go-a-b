package bandit

import (
	"math/rand"
)

// EpsilonGreedy represents the bandit data
type EpsilonGreedy struct {
	Epsilon float64   `json:"epsilon"`
	Counts  []int64   `json:"counts"`
	Values  []float64 `json:"values"`
	N       int       `json:"n"`
}

// SelectArm chooses an arm that exploits if the value is more than the epsilon
// threshold, and explore if the value is less than epsilon
func (b *EpsilonGreedy) SelectArm() (int, bool) {
	// Exploit
	if rand.Float64() > b.Epsilon {
		return max(b.Values...), true
	}
	// Explore
	return rand.Intn(len(b.Values)), false
}

// Update will update an arm with some reward value,
// e.g. click = 1, no click = 0
func (b *EpsilonGreedy) Update(chosenArm int, reward float64) {
	b.Counts[chosenArm]++
	n := float64(b.Counts[chosenArm])
	v := float64(b.Values[chosenArm])
	newValue := (v*(n-1) + reward) / n
	b.Values[chosenArm] = newValue
}

// SetValues sets the values to the input specified
func (b *EpsilonGreedy) SetValues(values []float64) {
	b.Values = values
}

// SetCounts sets the counts to the input specified
func (b *EpsilonGreedy) SetCounts(counts []int64) {
	b.Counts = counts
}

// New returns a pointer to the EpsilonGreedy struct
func New(nArms int, epsilonDecay float64) *EpsilonGreedy {
	return &EpsilonGreedy{
		N:       nArms,
		Epsilon: epsilonDecay,
		Values:  make([]float64, nArms),
		Counts:  make([]int64, nArms),
	}
}
