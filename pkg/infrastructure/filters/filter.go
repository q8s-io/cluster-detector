package filters

import (
	"strings"
)

// All filter interface
type Filter interface {
	Filter(event interface{}) (matched bool)
}

func GetValues(o []string) []string {
	if len(o) >= 1 {
		if len(o[0]) == 0 {
			return nil
		}
		return strings.Split(o[0], ",")
	}
	return nil
}
