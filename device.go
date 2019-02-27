package gecko

import "fmt"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 设备地址
type DeviceAddress struct {
	Group    string // 属组地址
	Private  string // 设备私有地址
	Internal string // 内部地址
}

// 获取Group/Private的联合地址
func (da DeviceAddress) GetUnionAddress() string {
	return MakeUnionAddress(da.Group, da.Private)
}

func (da DeviceAddress) String() string {
	return fmt.Sprintf(`{"group": "%s", "private": "%s", "internal": "%s"}`, da.Group, da.Private, da.Internal)
}

func (da DeviceAddress) IsValid() bool {
	return "" != da.Group && "" != da.Private
}

func (da DeviceAddress) Equals(to DeviceAddress) bool {
	return da.Group == to.Group &&
		da.Private == to.Private &&
		da.Internal == to.Internal
}

/////

// VirtualDevice是对硬件的抽象；
// 提供通讯地址和命名接口，以及支持的通讯协议
type VirtualDevice interface {
	Bundle
	// 设置设备地址
	setAddress(addr DeviceAddress)
	// 读取设备地址
	GetAddress() DeviceAddress
	// 设备名称
	setDisplayName(name string)
	GetDisplayName() string
	// 返回当前设备支持的通讯协议名称
	GetProtoName() string
	// 编码/解码
	setDecoder(decoder Decoder)
	GetDecoder() Decoder
	setEncoder(encoder Encoder)
	GetEncoder() Encoder
}

// 构建Union地址
func MakeUnionAddress(group, private string) string {
	return group + ":" + private
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
	displayName string
	address     DeviceAddress
	decoder     Decoder
	encoder     Encoder
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
func (dev *AbcInputDevice) setDisplayName(name string) {
	dev.displayName = name
}

func (dev *AbcInputDevice) GetDisplayName() string {
	return dev.displayName
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
	displayName string
	address     DeviceAddress
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

func (dev *AbcOutputDevice) setDisplayName(name string) {
	dev.displayName = name
}

func (dev *AbcOutputDevice) GetDisplayName() string {
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
