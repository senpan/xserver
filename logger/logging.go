package logger

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"
)

type XServerLogger struct {
}

func (x *XServerLogger) Debug(tag string, args ...interface{}) {
	fmt.Println(tag, args)
}

func (x *XServerLogger) Info(tag string, args ...interface{}) {
	fmt.Println(tag, args)
}

func (x *XServerLogger) Warn(tag string, args ...interface{}) {
	fmt.Println(tag, args)
}

func (x *XServerLogger) Error(tag string, args ...interface{}) {
	fmt.Println(tag, args)
}

func (x *XServerLogger) Fatal(tag string, args ...interface{}) {
	fmt.Println(tag, args)
}

func (x *XServerLogger) Debugf(tag string, format string, args ...interface{}) {
	fmt.Printf("tag:%s,"+format, tag, args)
}

func (x *XServerLogger) Infof(tag string, format string, args ...interface{}) {
	fmt.Printf("tag:%s,"+format, tag, args)
}

func (x *XServerLogger) Warnf(tag string, format string, args ...interface{}) {
	fmt.Printf("tag:%s,"+format, tag, args)
}

func (x *XServerLogger) Errorf(tag string, format string, args ...interface{}) {
	fmt.Printf("tag:%s,"+format, tag, args)
}

func (x *XServerLogger) Fatalf(tag string, format string, args ...interface{}) {
	fmt.Printf("tag:%s,"+format, tag, args)
}

func (x *XServerLogger) Close() {

}

type RocketMQLog struct {
}

func (s *RocketMQLog) Debug(msg string, fields map[string]interface{}) {
	GetLogger().Debugf("RocketMQLog", "%s,fields:%v", msg, fields)
}

func (s *RocketMQLog) Info(msg string, fields map[string]interface{}) {
	if strings.Contains(msg, "Stats In One Minute") {
		GetLogger().Infof("RocketMQLog", "[RocketMQ Stat] topic:%s,statsName:%s,sum:%s,tps:%s,avgpt:%s",
			fields["statsKey"], fields["statsName"],
			cast.ToString(fields["SUM"]), cast.ToString(fields["TPS"]), cast.ToString(fields["AVGPT"]))
		return
	}
	GetLogger().Debugf("RocketMQLog", "%s,fields:%v", msg, fields)
}

func (s *RocketMQLog) Warning(msg string, fields map[string]interface{}) {
	GetLogger().Warnf("RocketMQLog", "%s,fields:%v", msg, fields)
}

func (s *RocketMQLog) Error(msg string, fields map[string]interface{}) {
	GetLogger().Errorf("RocketMQLog", "%s,fields:%v", msg, fields)
}
func (s *RocketMQLog) Fatal(msg string, fields map[string]interface{}) {
	GetLogger().Fatalf("RocketMQLog", "%s,fields:%v", msg, fields)
}

func (s *RocketMQLog) Level(level string) {
}

func (s *RocketMQLog) OutputPath(path string) (err error) {
	return nil
}
