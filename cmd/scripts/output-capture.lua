---
--- 通过Lua执行一个截图指令,并返回截图结果文件
---

function process(args, frameStr)
    commands = { "ffmpeg",
                 "-i", "rtsp://user@password:ip",
                 "-t", "0.001",
                 "-f", "image2",
                 "-vframe", "1",
                 "-rtsp_transport", "tcp",
                 "/tmp/SNAPSHOT.png"
    }
    cmd = table.concat(commands, " ")
    print("正在执行命令: ", cmd)
    ret = os.execute(cmd)
    if nil == ret then
        return [[ { "status", "success", "file": "/tmp/SNAPSHOT.png" } ]], nil
    else
        return nil, "执行命令错误:" .. ret
    end
end