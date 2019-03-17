package gecko

import (
	"errors"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// VirtualDevice是对硬件的抽象；
type VirtualDevice interface {
	// 内部函数
	setUuid(uuid string)
	setName(name string)
	setDecoder(decoder Decoder)
	setEncoder(encoder Encoder)
	// 公开可访问函数
	GetUuid() string
	GetName() string
	GetDecoder() Decoder
	GetEncoder() Encoder
}

//// Input

// Input设备是表示向系统输入数据的设备
type InputDevice interface {
	VirtualDevice
	// 输入设备都具有一个Topic
	setTopic(topic string)
	GetTopic() string
	// 逻辑设备
	addLogicDevice(device LogicDevice) error
	GetLogicDevices() []LogicDevice
	// 监听设备的输入数据。如果设备发生错误，返回错误信息。
	Serve(ctx Context, deliverer InputDeliverer) error
}

////

// AbcInputDevice
type AbcInputDevice struct {
	InputDevice
	name    string
	uuid    string
	decoder Decoder
	encoder Encoder
	topic   string
	logics  map[string]LogicDevice
}

func (d *AbcInputDevice) setTopic(topic string) {
	d.topic = topic
}

func (d *AbcInputDevice) GetTopic() string {
	return d.topic
}

func (d *AbcInputDevice) setDecoder(decoder Decoder) {
	d.decoder = decoder
}

func (d *AbcInputDevice) GetDecoder() Decoder {
	return d.decoder
}

func (d *AbcInputDevice) setEncoder(encoder Encoder) {
	d.encoder = encoder
}

func (d *AbcInputDevice) GetEncoder() Encoder {
	return d.encoder
}

func (d *AbcInputDevice) setName(name string) {
	d.name = name
}

func (d *AbcInputDevice) GetName() string {
	return d.name
}

func (d *AbcInputDevice) setUuid(uuid string) {
	d.uuid = uuid
}

func (d *AbcInputDevice) GetUuid() string {
	return d.uuid
}

// 逻辑设备
func (d *AbcInputDevice) addLogicDevice(device LogicDevice) error {
	uuid := device.GetUuid()
	if _, ok := d.logics[uuid]; ok {
		return errors.New("LogicDevice uuid重复：" + uuid)
	} else {
		d.logics[uuid] = device
		return nil
	}
}

func (d *AbcInputDevice) GetLogicDevices() []LogicDevice {
	output := make([]LogicDevice, 0, len(d.logics))
	for _, dev := range d.logics {
		output = append(output, dev)
	}
	return output
}

func NewAbcInputDevice() *AbcInputDevice {
	return &AbcInputDevice{
		logics: make(map[string]LogicDevice),
	}
}

//// Output

// OutputDevice 是可交互的硬件的设备。它可以接收派发到此设备的事件，做出操作后，返回一个响应事件。
type OutputDevice interface {
	VirtualDevice
	// 设备对象接收控制事件；经设备驱动处理后，返回处理结果事件；
	Process(frame FramePacket, ctx Context) (FramePacket, error)
}

////

// AbcOutputDevice
type AbcOutputDevice struct {
	OutputDevice
	displayName string
	uuid        string
	decoder     Decoder
	encoder     Encoder
}

func (dev *AbcOutputDevice) setDecoder(decoder Decoder) {
	dev.decoder = decoder
}

func (dev *AbcOutputDevice) GetDecoder() Decoder {
	return dev.decoder
}

func (dev *AbcOutputDevice) setEncoder(encoder Encoder) {
	dev.encoder = encoder
}

func (dev *AbcOutputDevice) GetEncoder() Encoder {
	return dev.encoder
}

func (dev *AbcOutputDevice) setName(name string) {
	dev.displayName = name
}

func (dev *AbcOutputDevice) GetName() string {
	return dev.displayName
}

func (dev *AbcOutputDevice) setUuid(uuid string) {
	dev.uuid = uuid
}

func (dev *AbcOutputDevice) GetUuid() string {
	return dev.uuid
}

func NewAbcOutputDevice() *AbcOutputDevice {
	return new(AbcOutputDevice)
}
