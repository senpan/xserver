package core

type MQHandler func(topic string, data []byte, other ...string) error
