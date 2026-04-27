package healthcheck

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	defaultAttemptThresholdEnv = "NB_RELAY_HC_ATTEMPT_THRESHOLD"
)

func getAttemptThresholdFromEnv() int {
	if attemptThreshold := os.Getenv(defaultAttemptThresholdEnv); attemptThreshold != "" {
		threshold, err := strconv.Atoi(attemptThreshold)
		if err != nil || threshold <= 0 {
			log.Errorf("Failed to parse attempt threshold from environment variable \"%s\" should be a positive integer. Using default value", attemptThreshold)
			return defaultAttemptThreshold
		}
		return threshold
	}
	return defaultAttemptThreshold
}
