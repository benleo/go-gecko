package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// OutputDeliverer，用于向Output设备发送指令请求，并返回Output设备的处理结果。
// @param uuid 设备UUID地址
// @param message 指令数据包；其中数据包为系统内部Message格式，将由OutputDevice的Encoder编码器编码成字节数据；
// @return decodedMessage 响应数据格式；其中数据包为系统内部Message格式，当OutputDevice返回响应消息时，由Decoder解码成系统内部消息格式；
type OutputDeliverer func(uuid string, message *MessagePacket) (decodedMessage *MessagePacket, err error)

// @see OutputDeliverer
func (fn OutputDeliverer) Deliver(uuid string, data *MessagePacket) (decodedMessage *MessagePacket, err error) {
	return fn(uuid, data)
}

// OutputDevice 是可交互的硬件的设备。它可以接收派发到此设备的事件，做出操作后，返回一个响应事件。
type OutputDevice interface {
	VirtualDevice
	// 设备对象接收控制事件；经设备驱动处理后，返回处理结果事件；
	Process(encodedFrame FramePacket, ctx Context) (rawFrame FramePacket, err error)
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
