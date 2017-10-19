package bandit

import (
	"log"
	"testing"
)

func TestSum(t *testing.T) {
	testTable := []struct {
		in  []int
		out int
	}{
		{[]int{1, 2, 3, 4, 5}, 15},
		{[]int{-2, -1, 0, 1, 2}, 0},
		{[]int{10, 1, 20, 2, 30}, 63},
	}

	for _, v := range testTable {
		want := v.out
		got := sum(v.in...)
		if got != want {
			t.Errorf("want %v, got %v", want, got)
		}
	}
}

func TestMax(t *testing.T) {
	testTable := []struct {
		in  []float64
		out int
	}{
		{[]float64{-100, -50, 0, 50, 100}, 100},
		{[]float64{1, 0}, 1},
		{[]float64{1000, 1000}, 1000},
		{[]float64{-100, -1000}, -100},
	}

	for _, v := range testTable {
		want := v.out
		got := max(v.in...)

		if got != want {
			t.Errorf("want %v, got %v", want, got)
		}
	}
}

func TestNewBandit(t *testing.T) {
	nArms := 10
	epsilon := float64(0.1)

	bandit := New(nArms, epsilon)
	log.Printf("got bandit: %#v", bandit)

	wantN := nArms
	gotN := bandit.n
	if wantN != gotN {
		t.Errorf("want %v, got %v", wantN, gotN)
	}

	wantEpsilon := epsilon
	gotEpsilon := bandit.epsilon
	if wantEpsilon != gotEpsilon {
		t.Errorf("want %v, got %v", wantEpsilon, gotEpsilon)
	}

	arm, _ := bandit.chooseArm()
	wantArm := 0
	gotArm := arm
	if wantArm != gotArm {
		t.Errorf("want %v, got %v", wantArm, gotArm)
	}

	bandit.update(arm, 1)
	wantValues := float64(0)
	gotValues := float64(bandit.values[arm])
	log.Printf("got %#v", bandit)
	if wantValues != gotValues {
		t.Errorf("want %v, got %v", wantValues, gotValues)
	}
}

func TestPull(t *testing.T) {

	nArms := 5
	epsilon := float64(0.1)
	bandit := New(nArms, epsilon)
	exploitCount := 0
	for i := 0; i < 100000; i++ {
		arm, exploit := bandit.chooseArm()
		reward := 0
		if bernoulliArm() {
			reward = 1
		}
		if exploit == true {
			exploitCount += 1
		}
		bandit.update(arm, float64(reward))
	}

	log.Printf("got %#v:", bandit)
	log.Println("exploit_count", exploitCount)
	if 1 != 0 {
		t.Errorf("want %v, got %v", 1, 0)
	}
}
