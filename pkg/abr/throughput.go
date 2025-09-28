package abr

// ThroughputBased implements adaptive bitrate selection based on throughput
type ThroughputBased struct {
	// Add fields for storing historical data if needed
}

// NewThroughputBased creates a new throughput-based ABR algorithm
func NewThroughputBased() *ThroughputBased {
	return &ThroughputBased{}
}

// ChooseBitrate selects the appropriate bitrate based on available bandwidth
func (t *ThroughputBased) ChooseBitrate(bandwidth int, profiles []int) int {
	// Simple implementation: choose the highest bitrate that's <= 80% of bandwidth
	target := int(float64(bandwidth) * 0.8)
	
	var chosen int
	for _, profile := range profiles {
		if profile <= target {
			chosen = profile
		}
	}
	
	// If no profile fits, use the lowest one
	if chosen == 0 && len(profiles) > 0 {
		chosen = profiles[0]
		for _, profile := range profiles {
			if profile < chosen {
				chosen = profile
			}
		}
	}
	
	return chosen
}
