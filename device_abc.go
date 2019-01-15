package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 虚拟设备对象抽象实现
type AbcInteractiveDevice struct {
	InteractiveDevice
	displayName    string
	groupAddress   string
	privateAddress string
}

func (dev *AbcInteractiveDevice) setDisplayName(name string) {
	dev.displayName = name
}

func (dev *AbcInteractiveDevice) GetDisplayName() string {
	return dev.displayName
}

func (dev *AbcInteractiveDevice) setGroupAddress(addr string) {
	dev.groupAddress = addr
}

func (dev *AbcInteractiveDevice) GetGroupAddress() string {
	return dev.groupAddress
}

func (dev *AbcInteractiveDevice) setPrivateAddress(addr string) {
	dev.privateAddress = addr
}

func (dev *AbcInteractiveDevice) GetPrivateAddress() string {
	return dev.privateAddress
}

func (dev *AbcInteractiveDevice) GetUnionAddress() string {
	return "/" + dev.groupAddress + "/" + dev.privateAddress
}

func NewAbcInteractiveDevice() *AbcInteractiveDevice {
	return new(AbcInteractiveDevice)
}

//// InputDevice

type AbcInputDevice struct {
	InputDevice
	displayName    string
	groupAddress   string
	privateAddress string
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
	return "/" + dev.groupAddress + "/" + dev.privateAddress
}

func NewAbcInputDevice() *AbcInputDevice {
	return new(AbcInputDevice)
}

//// OutputDevice

type AbcOutputDevice struct {
	OutputDevice
	displayName    string
	groupAddress   string
	privateAddress string
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
	return "/" + dev.groupAddress + "/" + dev.privateAddress
}

func NewAbcOutputDevice() *AbcOutputDevice {
	return new(AbcOutputDevice)
}
