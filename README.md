
# Gecko概念

## 一、InputDevice/OutputDevice - 输入输出设备

在Gecko系统中，**InputDevice**和**OutputDevice**都是一个虚拟设备概念，它们都是代表一个真实具体设备，向Gecko系统输入、输出数据的设备代理；

在与硬件通讯时，InputDevice/OutputDevice可以是通过_UART/USB/RS232/RS485_等硬件通讯接口，可以与ID/IC读卡器、指纹机设备、人脸识别设备、门禁电磁开关等硬件设备进行数据交互；

在与软件通讯时，InputDevice/OutputDevice可以通过_UDP/TCP_等通讯协议，接收并响应通过网络接口通讯的设备；这些终端设备可能也是上述硬件设备；

在与物联网云通讯时，InputDevice/OutputDevice可以通过MQTT、HTTP等协议，接收远程硬件、第三方Web服务发送的控制事件；

### 虚拟设备应用场景

`InputDevice`由是指能够读取到真实硬件、外部系统控制事件等数据的虚拟输入端。在我们的实际项目中，InputDevice通过UDP通讯协议，作为UDP服务器，接收来自门禁主板主动上报的刷卡数据、二维码扫描数据。

### 1.1 Encoder/Decoder - 编码解码器

`Encoder`和`Decoder`作用于InputDevice和OutputDeivce。当数据在**向设备对象`输入`**和**设备对象`输出`**时，会执行设备对象配置的编码解码器接口来处理数据。



### 1.2 内置支持的设备

1. `TCP / UDP` 通讯协议的InputDevice / OutputDevice；
2. `SerialPort` 通讯协议的InputDevice / OutputDevice；

## 二、Interceptor - 拦截器

// TODO

## 三、Driver - 用户驱动

Driver-`用户驱动`，是实现`设备与设备之间联动`、`设备事件响应业务处理`的核心组件，它通常与Interceptor一起完成某种业务功能；

在Gecko中，事件输入由InputDevice触发，经过Interceptor过滤后，到过Driver处理。
Driver通过接收特定事件Topic的事件，使用内部数据库、业务方法等逻辑计算后，控制OutputDeliverer下一级输出设备作出操作。
最典型的例子是：Driver接收到门禁刷卡ID后，驱动门锁开关设备；

## 四、Plugin - 插件

// TODO 

## 五、Hook - 生命周期钩子

// TODO 
