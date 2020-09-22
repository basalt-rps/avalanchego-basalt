package peer_sampling

import (
	"fmt"
	"time"
)

// TODO redirect this somewhere nicer
func Trace(format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	fmt.Printf("%.3f ~~ %s\n", float64(time.Now().UnixNano())/1_000_000_000., text)
}
