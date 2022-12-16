package log

import (
	"strings"

	logger "github.com/senpan/xlogger"
	"github.com/spf13/cast"
)

type RocketMQLog struct {
}

func (s *RocketMQLog) Debug(msg string, fields map[string]interface{}) {
	logger.D("RocketMQLog", "%s,fields:%v", msg, fields)
}

func (s *RocketMQLog) Info(msg string, fields map[string]interface{}) {
	if strings.Contains(msg, "Stats In One Minute") {
		logger.I("RocketMQLog", "[RocketMQ Stat] topic:%s,statsName:%s,sum:%s,tps:%s,avgpt:%s",
			fields["statsKey"], fields["statsName"],
			cast.ToString(fields["SUM"]), cast.ToString(fields["TPS"]), cast.ToString(fields["AVGPT"]))
		return
	}
	logger.D("RocketMQLog", "%s,fields:%v", msg, fields)
}

func (s *RocketMQLog) Warning(msg string, fields map[string]interface{}) {
	logger.W("RocketMQLog", "%s,fields:%v", msg, fields)
}

func (s *RocketMQLog) Error(msg string, fields map[string]interface{}) {
	logger.E("RocketMQLog", "%s,fields:%v", msg, fields)
}
func (s *RocketMQLog) Fatal(msg string, fields map[string]interface{}) {
	logger.F("RocketMQLog", "%s,fields:%v", msg, fields)
}

func (s *RocketMQLog) Level(level string) {
}

func (s *RocketMQLog) OutputPath(path string) (err error) {
	return nil
}
