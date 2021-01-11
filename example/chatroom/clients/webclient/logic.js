let ws = new WebSocket("ws://localhost:20220");
ws.binaryType = 'arraybuffer';
// xhr.responseType = "arraybuffer";
// message code demo
const Enter_Room = 1001
const Chat_Message = 1002
const Leave_Room = 1003
const Join_Success = 2001

// send message demo
function SendMessage(objMsg) {
    let strMsg = JSON.stringify(objMsg)
    let arr = stringToByte(strMsg)
    ws.send(new Uint8Array(arr))
}

// onmessage demo
let OnMessage = function (data) {
    let bytes = new Uint8Array(data)
    // 解码成字符串
    let dataString = byteToString(bytes)
    let msgObject = JSON.parse(dataString)
    console.log(Join_Success)
    switch (msgObject.id) {
        case Join_Success:
            console.log("加入房间成功")
            show_prompt()
            break
        case Chat_Message:
            let msg = msgObject.content
            console.log("收到来自:" + msg.FromName + ",内容：" + msg.Content)
            let box = document.getElementById("pc-chat-box")
            let oneMsgBox = document.createElement("div")
            oneMsgBox.setAttribute("class", "msg-box")
            oneMsgBox.innerHTML =
                "<img class=\"msg-box-avatar\" src=\"default.jpg\">"
                + "<span class=\"msg-box-name\">" + msg.FromName + "：</span>"
                + "<span class=\"msg-box-text\">" + msg.Content + "</span>"
            box.append(oneMsgBox)
            box.scrollTop = box.scrollHeight;
            break
        default:
            console.error("miss message id::", msgObject.id)
    }
}

window.onload = function () {
    let input = document.getElementById("say")
    let sendBt = document.getElementById("send")

    function sendBox() {
        let content = input.value
        if (content.trim() === "") {
            alert("不能发送空消息")
            return null
        }
        let msg = {"Id": Chat_Message, "Content": {"FromName": This.name, "Content": content}}
        SendMessage(msg)
        input.value = ""
    }

    sendBt.onclick = sendBox
    document.body.addEventListener('keyup', function (e) {
        if (e.keyCode == '13') {
            sendBox()
        }
    })
}

/*JsonProtocol {
    Id int
    Content Any
}*/
//与服务端建立连接触发
ws.onopen = function () {
    console.log("与服务器成功建立连接")
    let data = {"id": Enter_Room, "content": {}}
    SendMessage(data)
    setInterval(function () {
        let ping = {"id": 0, content: {}}
        SendMessage(ping)
    }, 10000)
};
//服务端推送消息触发
ws.onmessage = function (ev) {
    OnMessage(ev.data);
};

//发生错误触发
ws.onerror = function () {
    console.log("连接错误")
};
//正常关闭触发
ws.onclose = function () {
    console.log("连接关闭");
};
let This = {}

function show_prompt() {
    let value = prompt('已进入聊天室，输入你的名字：', '匿名游客');
    if (value == null) {
        alert('您得告诉大家你是谁鸭！');
        show_prompt();
    } else if (value == '') {
        alert('姓名输入为空，请重新输入！');
        show_prompt();
    } else {
        This.name = value
    }
}

// string <=> bytes UTF-8
function stringToByte(str) {
    var bytes = new Array();
    var len, c;
    len = str.length;
    for (var i = 0; i < len; i++) {
        c = str.charCodeAt(i);
        if (c >= 0x010000 && c <= 0x10FFFF) {
            bytes.push(((c >> 18) & 0x07) | 0xF0);
            bytes.push(((c >> 12) & 0x3F) | 0x80);
            bytes.push(((c >> 6) & 0x3F) | 0x80);
            bytes.push((c & 0x3F) | 0x80);
        } else if (c >= 0x000800 && c <= 0x00FFFF) {
            bytes.push(((c >> 12) & 0x0F) | 0xE0);
            bytes.push(((c >> 6) & 0x3F) | 0x80);
            bytes.push((c & 0x3F) | 0x80);
        } else if (c >= 0x000080 && c <= 0x0007FF) {
            bytes.push(((c >> 6) & 0x1F) | 0xC0);
            bytes.push((c & 0x3F) | 0x80);
        } else {
            bytes.push(c & 0xFF);
        }
    }
    return bytes;


}


function byteToString(arr) {
    if (typeof arr === 'string') {
        return arr;
    }
    var str = '',
        _arr = arr;
    for (var i = 0; i < _arr.length; i++) {
        var one = _arr[i].toString(2),
            v = one.match(/^1+?(?=0)/);
        if (v && one.length == 8) {
            var bytesLength = v[0].length;
            var store = _arr[i].toString(2).slice(7 - bytesLength);
            for (var st = 1; st < bytesLength; st++) {
                store += _arr[st + i].toString(2).slice(2);
            }
            str += String.fromCharCode(parseInt(store, 2));
            i += bytesLength - 1;
        } else {
            str += String.fromCharCode(_arr[i]);
        }
    }
    return str;
}