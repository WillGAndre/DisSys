package common

import (
	"bufio"
	"log"
	"math"
	"os"

	"golang.org/x/exp/rand"
)

const LAMBDA = 2

/*
	Events bellow are generated assuming abstract time = 1.
	We define λ = 2 since by multiplying the generated timestamps
	by 60s -- 1min, we can conclude that 1ev -- 30s <=> 2ev -- 60s.
	We do this to be able to define real time values spaced out by event frequency.
*/
func PoissonProcessTimeToNextEvent() float64 {
	return (-math.Log(1-rand.Float64()) / LAMBDA)
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

// aux

func Insert(s []string, i int, v string) []string { // len(s) > 1
	s = append(s, "")
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}

func Contains(stack interface{}, needle interface{}) bool {
	switch stack := stack.(type) {
	case []string:
		for _, s := range stack {
			if s == needle {
				return true
			}
		}
	case []uint16:
		for _, s := range stack {
			if s == needle {
				return true
			}
		}
	}
	return false
}

func RemoveByIndex(s []string, i int) []string {
	return append(s[:i], s[i+1:]...)
}

func GetRandString(s int) string {
	alpha := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, s)
	for i := range b {
		b[i] = alpha[rand.Intn(len(alpha))]
	}
	return string(b)
}
