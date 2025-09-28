package player

import (
	"fmt"
	"time"
)

// Player represents a video player with ABR capabilities
type Player struct {
	manifestURL    string
	currentBitrate int
	bufferLevel    time.Duration
}

// New creates a new player instance
func New(manifestURL string) *Player {
	return &Player{
		manifestURL: manifestURL,
	}
}

// Start begins playback
func (p *Player) Start() error {
	fmt.Printf("Starting playback of: %s\n", p.manifestURL)
	// Implementation would fetch manifest, start downloading segments, etc.
	return nil
}

// SetBitrate changes the current bitrate
func (p *Player) SetBitrate(bitrate int) {
	p.currentBitrate = bitrate
	fmt.Printf("Switched to bitrate: %d\n", bitrate)
}

// GetBufferLevel returns the current buffer level
func (p *Player) GetBufferLevel() time.Duration {
	return p.bufferLevel
}

// GetCurrentBitrate returns the current bitrate
func (p *Player) GetCurrentBitrate() int {
	return p.currentBitrate
}
