package web

import (
	"testing"
)

func TestTicker(t *testing.T) {
	const (
		startMinute = 4  // 4 minutes after the hour
		spanMinute  = 15 // every 15 minutes
	)

	tick := firstTicker(startMinute, spanMinute)
	t.Logf("%v", tick.String())
	// next := time.Now().Add(tick)
	// if next.Minute() != startMinute {
	// 	t.Log(next.Minute(), startMinute)
	// 	t.Fatal()
	// }
}
