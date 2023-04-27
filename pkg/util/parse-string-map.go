package util

import (
	"encoding/json"
	"fmt"
)

func ParseStringifiedMap(stringifiedMap string) (map[string]string, error) {
	var stringMap map[string]string
	err := json.Unmarshal([]byte(stringifiedMap), &stringMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse NETWORKS_ETHERSCAN_URL, error: %v", err)
	}

	return stringMap, nil
}
