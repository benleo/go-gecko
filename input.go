package gecko

import (
	"errors"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// InputDeliverer，用于Input设备发起输入事件数据，并获取系统处理结果数据；
// @param topic 输入事件的Topic；
// @param frame 输入事件数据包；其中数据包为字节数据格式，将由InputDevice的Decoder解码成系统内部消息格式；
// @return frame 返回响应数据包；其中数据包将由系统内部调用InputDevice的Encoder编码器将Message编码成字节数据；
type InputDeliverer func(topic string, rawFrame FramePacket) (encodedFrame FramePacket, err error)

func (fn InputDeliverer) Deliver(topic string, rawFrame FramePacket) (encodedFrame FramePacket, err error) {
	return fn(topic, rawFrame)
}

// Input设备是表示向系统输入数据的设备
type InputDevice interface {
	VirtualDevice
	LifeCycle
	// 输入设备都具有一个Topic
	setTopic(topic string)
	GetTopic() string
	// 逻辑设备
	addLogic(device LogicDevice) error
	GetLogicList() []LogicDevice

	// Serve 函数是设备的监听服务函数。它被一个单独协程启动，并阻塞运行；
	// 如果此函数返回错误，系统终止所在协程。
	Serve(ctx Context, deliverer InputDeliverer) error
}

//// ABC

type AbcInputDevice struct {
	InputDevice
	name       string
	uuid       string
	decoder    Decoder
	encoder    Encoder
	topic      string
	namedLogic map[string]LogicDevice
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
func (d *AbcInputDevice) addLogic(device LogicDevice) error {
	uuid := device.GetUuid()
	if _, ok := d.namedLogic[uuid]; ok {
		return errors.New("LogicDevice uuid重复：" + uuid)
	} else {
		d.namedLogic[uuid] = device
		return nil
	}
}

func (d *AbcInputDevice) GetLogicList() []LogicDevice {
	output := make([]LogicDevice, 0, len(d.namedLogic))
	for _, dev := range d.namedLogic {
		output = append(output, dev)
	}
	return output
}

func NewAbcInputDevice() *AbcInputDevice {
	return &AbcInputDevice{
		namedLogic: make(map[string]LogicDevice),
	}
}
