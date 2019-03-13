package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

////

// LogicDevice，非常轻量级的逻辑设备；
// 它用于将输入数据，根据一定的逻辑关系，转换成另一个虚拟的只存在于逻辑关系的设备；
// 逻辑设备的应用场景是：市场上大部分控制门禁主板都至少包含1-4个门锁开关接口，1-2个读卡器接口。
// 在硬件上，门禁主板才是实际的设备，它们使用统一的TCP/UDP/RS485等协议来通讯；但其内部门锁开关不能直接映射到输入设备实体上，因为他们只存在于数据逻辑中。
type LogicDevice interface {
	Initialize
	// 内部函数
	setUuid(uuid string)
	setTopic(topic string)
	setName(name string)
	setMasterUuid(masterUuid string)
	// 公开可访问函数
	GetUuid() string
	GetName() string
	GetTopic() string
	GetMasterUuid() string
	// 检查是否符合逻辑设备的数据
	CheckIfMatch(json JSONPacket) bool
	// 转换输入的数据
	Transform(json JSONPacket) (newJson JSONPacket)
}

type AbcLogicDevice struct {
	LogicDevice
	name       string
	uuid       string
	topic      string
	masterUuid string
}

func (s *AbcLogicDevice) setUuid(uuid string) {
	s.uuid = uuid
}

func (s *AbcLogicDevice) setName(name string) {
	s.name = name
}

func (s *AbcLogicDevice) setTopic(topic string) {
	s.topic = topic
}

func (s *AbcLogicDevice) setMasterUuid(masterUuid string) {
	s.masterUuid = masterUuid
}

func (s *AbcLogicDevice) GetUuid() string {
	return s.uuid
}

func (s *AbcLogicDevice) GetName() string {
	return s.name
}

func (s *AbcLogicDevice) GetTopic() string {
	return s.topic
}

func (s *AbcLogicDevice) GetMasterUuid() string {
	return s.masterUuid
}

func NewAbcLogicDevice() *AbcLogicDevice {
	return new(AbcLogicDevice)
}
