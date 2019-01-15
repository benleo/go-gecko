package bundles

import (
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func UdpProtoPipelineFactory() (string, gecko.BundleFactory) {
	return "UdpProtoPipeline", func() interface{} {
		return NewUdpProtoPipeline()
	}
}

func NewUdpProtoPipeline() gecko.ProtoPipeline {
	return &UdpProtoPipeline{
		AbcProtoPipeline: gecko.NewAbcProtoPipeline(),
	}
}

// UDP通讯协议Pipeline
type UdpProtoPipeline struct {
	*gecko.AbcProtoPipeline
}

func (up *UdpProtoPipeline) GetProtoName() string {
	return "udp"
}

func (up *UdpProtoPipeline) OnInit(args map[string]interface{}, ctx gecko.Context) {
	// nop
}