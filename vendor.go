package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 设备品牌信息
type VendorInfo interface {
	// 返回设备品牌名
	VendorName() string
	// 返回设备描述信息
	Description() string
}

// 设备控制接口
type Executable interface {
	// 返回设备控制指令JSON格式字符串
	ExecCommand() string
}
