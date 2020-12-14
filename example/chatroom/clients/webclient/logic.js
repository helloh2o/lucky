let ws = new WebSocket("ws://localhost:20220");
ws.binaryType = 'arraybuffer';

// message code demo
const Enter_Room = 1001
const Chat_Message = 1002
const Leave_Room = 1003
const Join_Success = 2001

// send message demo
function SendMessage(objMsg) {
    let strMsg = JSON.stringify(objMsg)
    let arr = stringToUint8Array(strMsg)
    ws.send(new Uint8Array(arr))
}
// onmessage demo
let OnMessage = function (data) {
    let bytes = new Uint8Array(data)
    // 解码成字符串
    //let decodedString = String.fromCharCode.apply(null, bytes);
    // parse,转成json数据
    let dataString = Uint8ArrayToString(bytes)
    let msgObject = JSON.parse(dataString)
    console.log(msgObject.Id)
    console.log(Join_Success)
    switch (msgObject.id) {
        case Join_Success:
            console.log("加入房间成功")
            break
        case Chat_Message:
            let msg = msgObject.content
            console.log("收到来自:" + msg.FromName +",内容：" + msg.Content)
            break
        default:
            console.error("miss message id::", msgObject.id)
    }
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
    setInterval(function (){
        ws.send("ping")
    },10000)
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

function Uint8ArrayToString(fileData) {
    let dataString = "";
    for (let i = 0; i < fileData.length; i++) {
        dataString += String.fromCharCode(fileData[i]);
    }
    return dataString
}


function stringToUint8Array(str) {
    let arr = [];
    for (let i = 0, j = str.length; i < j; ++i) {
        arr.push(str.charCodeAt(i));
    }
    return arr
}