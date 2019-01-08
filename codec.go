package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 解码器
type Decoder func(bytes []byte) (map[string]interface{}, error)

// 编码器
type Encoder func(data map[string]interface{}) ([]byte, error)
