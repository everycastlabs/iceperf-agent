package acceptance_tests

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestMeasureThroughputOnMetered(t *testing.T) {
	err := loadEnv()
	assert.NoError(t, err)

	// FIXME use config.yaml instead?
	// API_KEY := os.Getenv("METERED_TURN_API_KEY")
	// driver := metered.Driver{}

	// specifications.ServerMeasurementsSpecification(t, driver, "throughput")
}
