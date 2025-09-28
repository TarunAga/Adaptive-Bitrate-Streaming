package main

import (
	"log"

	"github.com/TarunAga/adaptive-bitrate-streaming/pkg/abr"
	"github.com/TarunAga/adaptive-bitrate-streaming/pkg/manifest"
	"github.com/TarunAga/adaptive-bitrate-streaming/pkg/player"
)

func main() {
	// Load master manifest
	manifestURL := "http://localhost:8080/master.m3u8"
	m, err := manifest.Fetch(manifestURL)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize ABR algorithm
	algo := abr.NewThroughputBased()

	// Simulate playback
	p := player.New(manifestURL)
	err = p.Start()
	if err != nil {
		log.Fatal(err)
	}

	// Example of using ABR to select bitrate
	if len(m.Bitrates) > 0 {
		selectedBitrate := algo.ChooseBitrate(1500000, m.Bitrates)
		p.SetBitrate(selectedBitrate)
	}
}