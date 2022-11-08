package common

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// same seed --> deterministic behaviour
func TestGetRandomWord(t *testing.T) {
	lines := GetLines("../peergossip/engmix.txt")
	fmt.Println(lines)
	wdi := rand.Intn(len(lines))
	fmt.Println(lines[wdi])
}

func TestPoisson(t *testing.T) {
	//fmt.Println(PoissonProcessTimeToNextEvent()) // sleep time until next proc
	//fmt.Println(PoissonProcessEvents(1))         // Number of events (words selected) in 1 unit time

	var interval float64
	i := 1
	samples := 5
	for i < samples {
		interval += PoissonProcessTimeToNextEvent()
		fmt.Printf("%d event at: %f\n", i, interval*60)
		i += 1
	}
}

func TestPoissonTime(t *testing.T) {
	var ut float64
	timestamps := make([]float64, 0)
	i := 1

	// from sample, generate timestamps
	// 1ut --> 1 minute --> 60 seg
	for i < 5 {
		ut += PoissonProcessTimeToNextEvent()
		timestamps = append(timestamps, ut*60)
		i += 1
	}

	i = 0
	start := time.Now()
	// ---
	fmt.Printf("Start: %v\n", start)
	fmt.Printf("Timestamps: %v\n", timestamps)
	// ---
	for i < len(timestamps) {
		v := timestamps[i]
		if time.Since(start).Seconds() >= v {
			fmt.Println("Event trigger")
			i += 1
		} else {
			evtime := start.Add(time.Duration(v))
			time.Sleep(time.Since(evtime))
		}
	}
}
