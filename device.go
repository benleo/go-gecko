package gecko

import (
	"errors"
	"github.com/parkingwang/go-conf"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// VirtualDevice是对硬件的抽象；
// 提供通讯地址和命名接口，以及支持的通讯协议
type VirtualDevice interface {
	Bundle
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

////

// ShadowDevice，是InputDevice的直属下级，非常轻量级的影子设备，
// 它用于判断将InputDevice的输入数据是否为影子设备数据；
// 如果符合，则转换事件发送者的标记数据为影子设备的数据；
type ShadowDevice interface {
	Initialize
	// 内部函数
	setUuid(uuid string)
	setName(name string)
	setMasterUuid(masterUuid string)
	// 公开可访问函数
	GetUuid() string
	GetName() string
	GetMasterUuid() string
	// 检查是否符合影子设备的数据
	IsShadow(json JSONPacket) bool
	// 转换输入的数据
	TransformInput(topic string, json JSONPacket) (newTopic string, newJson JSONPacket)
	// 转换返回给设备的数据
	TransformOutput(json JSONPacket) (newJson JSONPacket)
}

type AbcShadowDevice struct {
	ShadowDevice
	name       string
	uuid       string
	masterUuid string
}

func (s *AbcShadowDevice) setUuid(uuid string) {
	s.uuid = uuid
}

func (s *AbcShadowDevice) setName(name string) {
	s.name = name
}

func (s *AbcShadowDevice) setMasterUuid(masterUuid string) {
	s.masterUuid = masterUuid
}

func (s *AbcShadowDevice) GetUuid() string {
	return s.uuid
}

func (s *AbcShadowDevice) GetName() string {
	return s.name
}

func (s *AbcShadowDevice) GetMasterUuid() string {
	return s.masterUuid
}

func NewAbcShadowDevice() *AbcShadowDevice {
	return new(AbcShadowDevice)
}

//// Input

// Input设备是表示向系统输入数据的设备
type InputDevice interface {
	VirtualDevice
	// 输入设备都具有一个Topic
	setTopic(topic string)
	GetTopic() string
	// 影子设备
	addShadowDevice(device ShadowDevice) error
	GetShadowDevices() []ShadowDevice
	// 监听设备的输入数据。如果设备发生错误，返回错误信息。
	Serve(ctx Context, deliverer InputDeliverer) error
}

////

// AbcInputDevice
type AbcInputDevice struct {
	InputDevice
	args    *cfg.Config
	ctx     Context
	name    string
	uuid    string
	decoder Decoder
	encoder Encoder
	topic   string
	devices map[string]ShadowDevice
}

func (d *AbcInputDevice) OnInit(args *cfg.Config, ctx Context) {
	d.args = args
	d.ctx = ctx
}

func (d *AbcInputDevice) GetInitArgs() *cfg.Config {
	return d.args
}

func (d *AbcInputDevice) GetInitContext() Context {
	return d.ctx
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

// 影子设备
func (d *AbcInputDevice) addShadowDevice(device ShadowDevice) error {
	uuid := device.GetUuid()
	if _, ok := d.devices[uuid]; ok {
		return errors.New("shadow device uuid重复：" + uuid)
	} else {
		d.devices[uuid] = device
		return nil
	}
}

func (d *AbcInputDevice) GetShadowDevices() []ShadowDevice {
	output := make([]ShadowDevice, 0, len(d.devices))
	for _, dev := range d.devices {
		output = append(output, dev)
	}
	return output
}

func NewAbcInputDevice() *AbcInputDevice {
	return &AbcInputDevice{
		devices: make(map[string]ShadowDevice),
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
	args        *cfg.Config
	ctx         Context
	displayName string
	uuid        string
	decoder     Decoder
	encoder     Encoder
}

func (dev *AbcOutputDevice) OnInit(args *cfg.Config, ctx Context) {
	dev.args = args
	dev.ctx = ctx
}

func (dev *AbcOutputDevice) GetInitArgs() *cfg.Config {
	return dev.args
}

func (dev *AbcOutputDevice) GetInitContext() Context {
	return dev.ctx
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
