package common

import (
	"fmt"
	"testing"
)

func TestRand(t *testing.T) {
	fmt.Println(GenRandom())
}

func TestPoisson(t *testing.T) {
	fmt.Println(PoissonProcessTimeToNextEvent()) // sleep time until next proc
	fmt.Println(PoissonProcessEvents(1))         // Number of events (words selected) in 1 unit time
}
