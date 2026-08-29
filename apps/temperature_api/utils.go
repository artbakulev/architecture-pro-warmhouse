package main

import "math/rand"

func randomFloat(from float64, to float64) float64 {
	return from + rand.Float64()*(to-from)
}

func randomInt(from int, to int) int {
	return int(randomFloat(float64(from), float64(to)))
}
