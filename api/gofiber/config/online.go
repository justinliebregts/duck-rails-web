package config

import (
	"os"
	"strconv"
)

func RetrieveHeartBeatFrequencyAndLifeCycle() (freq int, lifecycle int) {
	freqStr := os.Getenv("ONLINE_FREQUENCY")
	lifecycleStr := os.Getenv("ONLINE_LIFE_CYCLE")

	freq, _ = strconv.Atoi(freqStr)
	lifecycle, _ = strconv.Atoi(lifecycleStr)
	return
}
