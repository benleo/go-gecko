package bundles

import "net"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type UdpPack struct {
	Data    []byte
	Address *net.UDPAddr
}
