package random

import (
	"math/rand"
	"time"
)

func Duration(maxMilliSeconds int) time.Duration {
	// No need to seed if you're using rand.NewSource
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return time.Duration(r.Intn(maxMilliSeconds)) * time.Millisecond
}
