package bandit

import (
	"log"
	"math"
	"math/rand"
	"time"
)

// Bandit represents the bandit data
type Bandit struct {
	epsilon float64
	counts  []int
	values  []float64
	n       int
}

func (b *Bandit) chooseArm() (int, bool) {
	if rand.Float64() > b.epsilon {
		// Exploit
		return max(b.values...), true
	}
	// Explore
	return rand.Intn(len(b.values)), false
}

// update will update an arm with some reward value,
// e.g. click = 1, no click = 0
func (b *Bandit) update(arm int, reward float64) {
	b.counts[arm] = b.counts[arm] + 1
	n := float64(b.counts[arm])
	v := float64(b.values[arm])

	//	newValue := ((n-1)/n)*v + (1/n)*reward

	newValue := (v*(n-1) + reward) / n
	b.values[arm] = newValue
}

func sum(values ...int) int {
	total := 0
	for _, v := range values {
		total += v
	}
	return total
}

func max(values ...float64) int {
	value := math.Inf(-1)
	for _, v := range values {
		if float64(v) > float64(value) {
			value = float64(v)
		}
	}
	return int(value)
}

func New(nArms int, epsilonDecay float64) *Bandit {
	return &Bandit{
		epsilon: epsilonDecay,
		n:       nArms,
		values:  make([]float64, nArms),
		counts:  make([]int, nArms),
	}
}

func (b *Bandit) SetValues(values []float64) {
	b.values = values
}

func (b *Bandit) SetCounts(counts []int) {
	b.counts = counts
}

func bernoulliArm() bool {
	return rand.Float32() < 0.5
}

func main() {
	rand.Seed(time.Now().UnixNano())
	bandit := New(5, 0.1)
	arm, _ := bandit.chooseArm()

	log.Println("arm:", arm)
	log.Println("randfloat", rand.Float32())
}
