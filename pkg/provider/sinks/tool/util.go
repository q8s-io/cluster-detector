package tool

import (
	"strings"

	v1 "k8s.io/api/core/v1"
)

func GetLevel(level string) int {
	score := 0
	switch level {
	case v1.EventTypeWarning:
		score += 2
	case v1.EventTypeNormal:
		score += 1
	default:
		//score will remain 0
	}
	return score
}

func GetValues(o string) []string {
	if o == "" {
		return nil
	}
	return strings.Split(o, ",")
}
