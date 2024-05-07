package lancelog

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestLog(t *testing.T) {

	logrus.Info("test logrus")

	SetLevel(DebugLevel)

	Trace("trace test")
	Debug("debug test")
	Info("info test")
	Warn("warn test")
	Error("error test")
	//Fatal("fatal test")

	WithFields(Fields{"hello": "world"}).Info("with fields test")

}
