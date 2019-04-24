---
--- 通过Lua执行一个截图指令,并返回截图结果文件
---

function outputMain(args, frameStr)
    commands = { "ffmpeg",
                 "-i", "rtsp://user@password:ip",
                 "-t", "0.001",
                 "-f", "image2",
                 "-vframe", "1",
                 "-rtsp_transport", "tcp",
                 "/tmp/SNAPSHOT.png"
    }
    cmd = table.concat(commands, " ")
    return [[ { "status", "success", "cmd": "]] .. cmd .. [[" } ]], nil
end