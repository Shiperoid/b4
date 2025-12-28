package discovery

import (
	"fmt"

	"github.com/daniellavrushin/b4/log"
)

func (ds *DiscoverySuite) getOptimalTTL() uint8 {
	if ds.optimalTTL > 0 {
		return ds.optimalTTL
	}

	base := baseConfig()
	base.Faking.SNI = true
	base.Faking.Strategy = "pastseq"
	base.Faking.SeqOffset = 10000
	base.Faking.SNISeqLength = 1
	base.Faking.SNIType = ds.bestPayload
	base.Fragmentation.Strategy = "combo"
	base.Fragmentation.SNIPosition = 1

	tmpPreset := ConfigPreset{Name: "ttl-probe", Config: base}
	ds.optimalTTL, _ = ds.findOptimalTTL(tmpPreset)

	if ds.optimalTTL == 0 {
		ds.optimalTTL = 8
	}

	return ds.optimalTTL
}
func (ds *DiscoverySuite) findOptimalTTL(basePreset ConfigPreset) (uint8, float64) {
	var bestTTL uint8
	var bestSpeed float64
	low, high := uint8(1), uint8(32)

	log.DiscoveryLogf("Binary search for minimum working TTL (range %d-%d)", low, high)

	for low < high {
		mid := (low + high) / 2

		preset := basePreset
		preset.Name = fmt.Sprintf("ttl-search-%d", mid)
		preset.Config.Faking.TTL = mid

		result := ds.testPresetWithBestPayload(preset)
		ds.storeResult(preset, result)

		if result.Status == CheckStatusComplete {
			bestTTL = mid
			bestSpeed = result.Speed
			high = mid
			log.DiscoveryLogf("  TTL %d: SUCCESS (%.2f KB/s)", mid, result.Speed/1024)
		} else {
			low = mid + 1
			log.Tracef("  TTL %d: FAILED", mid)
		}
	}

	if bestTTL > 0 {
		log.DiscoveryLogf("Minimum working TTL: %d (%.2f KB/s)", bestTTL, bestSpeed/1024)
		ds.optimalTTL = bestTTL
	}
	return bestTTL, bestSpeed
}
