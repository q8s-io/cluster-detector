package pod

import (
	"log"
	"testing"
	"time"
)

func TestJob(t *testing.T) {

	updatepodProbeTimer := time.NewTicker(time.Second * time.Duration(2))
	i := 0
	for {
		log.Print(i)
		i++
		<-updatepodProbeTimer.C
	}
}
