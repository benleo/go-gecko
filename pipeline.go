package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// ProtoPipeline 是可以处理一类设备通讯协议的管理类。
// 在功能实现中，Pipeline应当承担相应通讯协议的通讯底层，尽量将Device轻量化；
type ProtoPipeline interface {
	Bundle

	// 返回当前支持的通讯协议名称
	GetProtoName() string

	// 监听数据输入通道
	RegisterInputChannel(groupAddress, privateAddress string, channel PipelineChannel)

	// 监听数据输出通道；当数据
	RegisterOutputChannel(groupAddress, privateAddress string, channel PipelineChannel)

	// 添加设备对象。如果设备对象的地址重复，返回False。
	register(device VirtualDevice) bool
}

// 根据指定协议名，返回指定协议的Pipeline
type ProtoPipelineSelector func(proto string) (ProtoPipeline, bool)

// 扩展函数
func (s ProtoPipelineSelector) Find(proto string) (ProtoPipeline, bool) {
	return s(proto)
}

// Pipeline通道
type PipelineChannel interface {
	// 当数据发向注册的目标设备时，此函数回调
	OnReceive(frame PacketFrame)
	// 当设备
	OnSend(frame PacketFrame)
}

////

// ProtoPipeline抽象实现类
type AbcProtoPipeline struct {
	ProtoPipeline
	inputDevices  map[string]InputDevice
	outputDevices map[string]OutputDevice
}

func NewAbcProtoPipeline() *AbcProtoPipeline {
	return &AbcProtoPipeline{
		inputDevices:  make(map[string]InputDevice),
		outputDevices: make(map[string]OutputDevice),
	}
}

func (ap *AbcProtoPipeline) register(device VirtualDevice) bool {
	addr := device.GetUnionAddress()
	if in, ok := device.(InputDevice); ok {
		if _, has := ap.inputDevices[addr]; has {
			return false
		} else {
			ap.inputDevices[addr] = in
			return true
		}
	} else if out, ok := device.(OutputDevice); ok {
		if _, has := ap.inputDevices[addr]; has {
			return false
		} else {
			ap.outputDevices[addr] = out
			return true
		}
	}
	return false
}
