package utility

import (
	"math/rand"
	"time"
)

func ShuffleAndAdd[T any](addition T, array []T) (pos int, res []T) {
	// Seed the random function
	rand.Seed(time.Now().UnixNano())
	// Shuffle incoming array to get more uniform txn distribution
	rand.Shuffle(len(array), func(i, j int) { array[i], array[j] = array[j], array[i] })
	// Randomly choose the position of the true txn
	pos = rand.Intn(len(array))

	// Add the txn to the required position in the array
	res = append(res, array[:pos]...)
	res = append(res, addition)
	res = append(res, array[pos:]...)
	return
}
