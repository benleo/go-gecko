package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 硬件抽象，提供通讯地址和命名接口，以及支持的通讯协议
type VirtualDevice interface {
	Bundle
	// 属组地址
	setGroupAddress(addr string)
	GetGroupAddress() string
	// 设置设备私有地址
	setPrivateAddress(addr string)
	GetPrivateAddress() string
	// 获取设备地址，由 /{GroupAddress}/{PrivateAddress} 组成。
	GetUnionAddress() string
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
	displayName    string
	groupAddress   string
	privateAddress string
	decoder        Decoder
	encoder        Encoder
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

func (dev *AbcInputDevice) setGroupAddress(addr string) {
	dev.groupAddress = addr
}

func (dev *AbcInputDevice) GetGroupAddress() string {
	return dev.groupAddress
}

func (dev *AbcInputDevice) setPrivateAddress(addr string) {
	dev.privateAddress = addr
}

func (dev *AbcInputDevice) GetPrivateAddress() string {
	return dev.privateAddress
}

func (dev *AbcInputDevice) GetUnionAddress() string {
	return MakeUnionAddress(dev.groupAddress, dev.privateAddress)
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
	displayName    string
	groupAddress   string
	privateAddress string
	decoder        Decoder
	encoder        Encoder
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

func (dev *AbcOutputDevice) setGroupAddress(addr string) {
	dev.groupAddress = addr
}

func (dev *AbcOutputDevice) GetGroupAddress() string {
	return dev.groupAddress
}

func (dev *AbcOutputDevice) setPrivateAddress(addr string) {
	dev.privateAddress = addr
}

func (dev *AbcOutputDevice) GetPrivateAddress() string {
	return dev.privateAddress
}

func (dev *AbcOutputDevice) GetUnionAddress() string {
	return MakeUnionAddress(dev.groupAddress, dev.privateAddress)
}

func NewAbcOutputDevice() *AbcOutputDevice {
	return new(AbcOutputDevice)
}
