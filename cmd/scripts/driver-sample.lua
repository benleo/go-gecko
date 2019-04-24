--[[
    Driver脚本入口函数
    @Param args 配置文件定义的参数列表
    @Param inbound 事件输入参数
    @Param deliverFn Deliver函数，原型为： function(targetUuid, payloadTable) (respTable, error)
    @Return 返回两个参数：
        1. Table 处理结果；
        2. Error 错误；
]]--

function driverMain(args, inbound, deliverFn)
    uuid = args["targetUuid"]
    print("Target UUID: " .. uuid)
    print(type(deliverFn))
    ret, err = deliverFn(uuid, { foo = "bar" })
    print(ret)
    print(err)
    return ret, err
end