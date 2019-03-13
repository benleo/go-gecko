
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

### 1.3 LogicDevice - 逻辑设备

LogicDevice，非常轻量级的逻辑设备，依附于InputDeivce的具体通讯能力实现，只在数据层面转换数据包与设备的关联关系；

它用于将输入数据，根据一定的逻辑关系，转换成另一个只存在于数据逻辑关系上的设备；

它的应用场景是：市场上大部分控制门禁主板都至少包含1-4个门锁开关接口，1-2个读卡器接口。在硬件上，门禁主板才是实际的设备，它们使用统一的TCP/UDP/RS485等协议来通讯；但其内部门锁开关不能直接映射到输入设备实体上，因为他们只存在于数据逻辑中。

以真实项目举例：使用UDP/TCP通讯的门禁主板，在设定`NetworkInputDevice`来监听上报数据时，可以根据门禁主板的设计，为其设置几个逻辑设备：

1. `SW#0` 门禁主板消防联动开关；
1. `SW#1` 门禁主板1号开关；
1. `SW#2` 门禁主板2号开关；
1. `SW#3` 门禁主板3号开关；
1. `SW#4` 门禁主板4号开关；
1. `READER#1` 门禁主板1号RFID读头；
1. `READER#2` 门禁主板2号RFID读头；

这样设置后，NetworkInputDevice相当于一个网络通讯代理；当上报数据时，经过逻辑设备处理，内部系统会接收到事件时，可以按照独立设备上报事件的流程来处理。

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
