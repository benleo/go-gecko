package gecko

import (
	"fmt"
	"github.com/parkingwang/go-conf"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 设备地址
type DeviceAddress struct {
	UUID    string // 设备唯一ID
	Group   string // 属组地址
	Private string // 设备私有地址
}

func (da DeviceAddress) String() string {
	return fmt.Sprintf(`{"uuid": "%s", group": "%s", "private": "%s"}`, da.UUID, da.Group, da.Private)
}

func (da DeviceAddress) IsValid() bool {
	return da.UUID != "" && "" != da.Group && "" != da.Private
}

func (da DeviceAddress) Equals(to DeviceAddress) bool {
	return da.Group == to.Group &&
		da.Private == to.Private
}

/////

// VirtualDevice是对硬件的抽象；
// 提供通讯地址和命名接口，以及支持的通讯协议
type VirtualDevice interface {
	Bundle
	// 内部函数
	setAddress(addr DeviceAddress)
	setName(name string)
	setDecoder(decoder Decoder)
	setEncoder(encoder Encoder)
	// 公开可访问函数
	GetAddress() DeviceAddress
	GetName() string
	GetDecoder() Decoder
	GetEncoder() Encoder
}

//// Input

// Input设备是表示向系统输入数据的设备
type InputDevice interface {
	VirtualDevice
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
	address DeviceAddress
	decoder Decoder
	encoder Encoder
}

func (dev *AbcInputDevice) OnInit(args *cfg.Config, ctx Context) {
	dev.args = args
	dev.ctx = ctx
}

func (dev *AbcInputDevice) GetInitArgs() *cfg.Config {
	return dev.args
}

func (dev *AbcInputDevice) GetInitContext() Context {
	return dev.ctx
}

func (dev *AbcInputDevice) setDecoder(decoder Decoder) {
	dev.decoder = decoder
}

func (dev *AbcInputDevice) GetDecoder() Decoder {
	return dev.decoder
}

func (dev *AbcInputDevice) setEncoder(encoder Encoder) {
	dev.encoder = encoder
}

func (dev *AbcInputDevice) GetEncoder() Encoder {
	return dev.encoder
}

func (dev *AbcInputDevice) setName(name string) {
	dev.name = name
}

func (dev *AbcInputDevice) GetName() string {
	return dev.name
}

func (dev *AbcInputDevice) setAddress(addr DeviceAddress) {
	dev.address = addr
}

func (dev *AbcInputDevice) GetAddress() DeviceAddress {
	return dev.address
}

func NewAbcInputDevice() *AbcInputDevice {
	return new(AbcInputDevice)
}

//// Output

// OutputDevice 是可交互的硬件的设备。它可以接收派发到此设备的事件，做出操作后，返回一个响应事件。
type OutputDevice interface {
	VirtualDevice
	// 设备对象接收控制事件；经设备驱动处理后，返回处理结果事件；
	Process(frame PacketFrame, ctx Context) (PacketFrame, error)
}

////

// AbcOutputDevice
type AbcOutputDevice struct {
	OutputDevice
	args        *cfg.Config
	ctx         Context
	displayName string
	address     DeviceAddress
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

func (dev *AbcOutputDevice) setAddress(addr DeviceAddress) {
	dev.address = addr
}

func (dev *AbcOutputDevice) GetAddress() DeviceAddress {
	return dev.address
}

func NewAbcOutputDevice() *AbcOutputDevice {
	return new(AbcOutputDevice)
}
