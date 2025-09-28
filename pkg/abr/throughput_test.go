package abr

import "testing"

func TestChooseBitrate(t *testing.T) {
	algo := NewThroughputBased()
	bw := 1500000 // 1.5 Mbps
	profiles := []int{300000, 600000, 1200000}
	got := algo.ChooseBitrate(bw, profiles)
	want := 1200000
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}