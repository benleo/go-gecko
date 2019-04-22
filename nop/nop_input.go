package nop

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

import (
	"fmt"
	"github.com/yoojia/go-gecko/v2"
	"github.com/yoojia/go-value"
	"time"
)

func NopInputDeviceFactory() (string, gecko.Factory) {
	return "NopInputDevice", func() interface{} {
		return NewNopInputDevice()
	}
}

func NewNopInputDevice() *NopInputDevice {
	return &NopInputDevice{
		AbcInputDevice: gecko.NewAbcInputDevice(),
	}
}

// Nop客户端输入设备
type NopInputDevice struct {
	*gecko.AbcInputDevice
	gecko.Initial
	ticker *time.Ticker
}

func (d *NopInputDevice) OnInit(config map[string]interface{}, ctx gecko.Context) {
	d.ticker = time.NewTicker(value.Of(config["period"]).DurationOfDefault(time.Second))
}

func (d *NopInputDevice) OnStart(ctx gecko.Context) {

}

func (d *NopInputDevice) OnStop(ctx gecko.Context) {
	d.ticker.Stop()
}

func (d *NopInputDevice) Serve(ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	for t := range d.ticker.C {
		out, err := deliverer.Deliver(d.GetTopic(), []byte(fmt.Sprintf(`{"timestamp": %d}`, t.UnixNano())))
		if nil != err {
			log.Error(err)
		} else {
			log.Debug(out)
		}
	}
	return nil
}

func (d *NopInputDevice) VendorName() string {
	return "GoGecko/Nop/Input"
}

func (d *NopInputDevice) Description() string {
	return `定时返回数据的Input`
}
