
# Gecko概念

## InputDevice - 输入设备

在Gecko的设计中，InputDevice和OutputDevice都是一个虚拟概念的设备，它们都是代表`(represent)`一个真实具体设备向Gecko系统输入、输出数据；

`InputDevice`由是指能够读取到具体硬件设备、外部系统控制事件等的虚拟输入端。
在硬件方面，InputDevice可以是通过UART/USB/RS232/RS485等硬件通讯方式，与ID卡读卡器、指纹机、人脸识别设备、门禁开关等设备进行数据交互；
在软件方面，InputDevice可以通过UDP/TCP等方式，接收和返回通过网络接口通讯的设备；这些设备可能也是上述硬件设备；
在网络通讯方面来讲，InputDevice可以通过MQTT、HTTP等方式，接收远程硬件、第三方服务发送的控制事件；

## Interceptor - 拦截器

// TODO

## Driver - 用户驱动

Driver-`用户驱动`，是实现`设备与设备之间联动`、`设备事件响应业务处理`的核心组件，它通常与Interceptor一起完成某种业务功能；

在Gecko中，事件输入由InputDevice触发，经过Interceptor过滤后，到过Driver处理。
Driver通过接收特定事件Topic的事件，使用内部数据库、业务方法等逻辑计算后，控制OutputDeliverer下一级输出设备作出操作。
最典型的例子是：Driver接收到门禁刷卡ID后，驱动门锁开关设备；

## OutputDevice - 输出设备

// TODO

## Encoder/Decoder - 编码解码器

// TODO 

## Plugin - 插件

// TODO 

## Hook - 生命周期钩子

// TODO 

# 数据结构说明

## PacketFrame

// TODO 

## PacketMap

// TODO 