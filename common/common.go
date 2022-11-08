package common

import (
	"bufio"
	"log"
	"math"
	"os"

	"golang.org/x/exp/rand"
)

const LAMBDA = 2

func PoissonProcessTimeToNextEvent() float64 {
	return (-math.Log(1-rand.Float64()) / LAMBDA)
}

// time -> unit
// 1 --> 1 min --> 60 seg
func PoissonProcessEvents(time float64) float64 {
	var n float64
	n = 0
	p := math.Exp(-LAMBDA * time)
	s := p
	u := rand.Float64()
	for u > s {
		n += 1
		p = (p * LAMBDA) / n
		s = s + p
	}
	return n
}

func GetLines(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	res := []string{}

	for scanner.Scan() {
		res = append(res, scanner.Text())
	}

	return res
}
