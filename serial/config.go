package serial

import (
	"github.com/parkingwang/go-conf"
	"github.com/tarm/serial"
	"strings"
	"time"
)

func getSerialConfig(config *cfg.Config, timeout time.Duration) *serial.Config {
	parity := serial.ParityNone
	switch strings.ToUpper(config.MustString("parity")) {
	case "N", "NONE":
		parity = serial.ParityNone
	case "O", "ODD":
		parity = serial.ParityOdd
	case "E", "EVEN":
		parity = serial.ParityEven
	case "M", "MARK":
		parity = serial.ParityMark
	case "S", "SPACE":
		parity = serial.ParitySpace

	default:
		parity = serial.ParityNone
	}
	return &serial.Config{
		Name:        config.MustString("serialPort"),
		Baud:        int(config.MustInt64("baudRate")),
		Size:        byte(config.MustInt64("dataBit")),
		Parity:      parity,
		StopBits:    serial.StopBits(config.MustInt64("stopBits")),
		ReadTimeout: timeout,
	}
}
