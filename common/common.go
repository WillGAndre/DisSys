package common

import (
	"bufio"
	"log"
	"math"
	"os"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

const LAMBDA = 1

func PoissonProcessTimeToNextEvent() float64 {
	return (-math.Log(1-rand.Float64()) / LAMBDA)
}

// time -> unit
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

func GenRandom() int64 {
	source := rand.NewSource(uint64(time.Now().UnixNano()))
	p := distuv.Poisson{
		Lambda: LAMBDA,
		Src:    source,
	}
	return int64(p.Rand())
}

func GetLines() []string {
	f, err := os.Open("engmix.txt")
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
