package bandit

import (
	"log"
	"math"
	"math/rand"
	"time"
)

type Bandit struct {
	epsilon float64
	counts  []int
	values  []float64
	n       int
}

func (b *Bandit) chooseArm() (int, bool) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	if r1.Float64() > b.epsilon {
		// Exploit
		return max(b.values...), true
	}
	// Explore
	s2 := rand.NewSource(time.Now().UnixNano())
	r2 := rand.New(s2)
	return r2.Intn(len(b.values)), false
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

func bernoulliArm() bool {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Float32() < 0.5
}

func main() {

	bandit := New(5, 0.1)
	arm, _ := bandit.chooseArm()

	log.Println("arm:", arm)
	log.Println("randfloat", rand.Float32())

}
