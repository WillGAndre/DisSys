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

// Insert function, assuming len(s) > 1.
// Since the index to insert (order within the queue) is discovered by the multicast
// module, we simply create a new position for insertion, append the rhs (right-hand-side) of the queue
// to the lfs (left-hand-side) (of the queue), s[i+1:] & s[i:] respectively, and then insert our value.
func Insert(s []string, i int, v string) []string {
	s = append(s, "")
	copy(s[i+1:], s[i:])
	s[i] = v
	return s
}

// "loose" contains function
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

// removes a value by index, used by the multicast queue to remove msgs.
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
