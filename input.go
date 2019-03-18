package gecko

import (
	"errors"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// InputDeliverer，用于Input设备发起输入事件数据，并获取系统处理结果数据；
// @param topic 输入事件的Topic；
// @param frame 输入事件数据包；
type InputDeliverer func(topic string, frame FramePacket) (FramePacket, error)

func (fn InputDeliverer) Deliver(topic string, frame FramePacket) (FramePacket, error) {
	return fn(topic, frame)
}

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
func (d *AbcInputDevice) addLogicDevice(device LogicDevice) error {
	uuid := device.GetUuid()
	if _, ok := d.namedLogic[uuid]; ok {
		return errors.New("LogicDevice uuid重复：" + uuid)
	} else {
		d.namedLogic[uuid] = device
		return nil
	}
}

func (d *AbcInputDevice) GetLogicDevices() []LogicDevice {
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
