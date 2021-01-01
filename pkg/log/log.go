package log

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type LogMess struct {
	Namespace string
	Name      string
	Kind      string
	Type      string
	Time      time.Time
	Info      interface{}
}

func PrintLog(obj LogMess) {
	mes, err := json.Marshal(obj)
	if err != nil {
		logrus.Infof("marshal json error: %v\n", err.Error())
		return
	}
	fmt.Printf("%v\n", string(mes))
}
