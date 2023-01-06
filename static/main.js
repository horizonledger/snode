console.log("ready");

let socket = new WebSocket("ws://127.0.0.1:8080/ws");
console.log("Attempting Connection...");

socket.onopen = () => {
    console.log("Successfully Connected");
    socket.send("handshake")
};

socket.onclose = event => {
    console.log("Socket Closed Connection: ", event);
    socket.send("Client Closed!")
};

socket.onmessage = function (e) {
    console.log("Received: " + e.data);
    try {
        let msg = JSON.parse(e.data);
        console.log("msg " + msg);
        console.log("value: " + msg.value);
        console.log("type: " + msg.type);

        if (msg.type == "chat") {
            document.getElementById("log").textContent += "\n" + msg.value;
        } else if (msg.type == "uuid") {
            document.getElementById("uuid").textContent += "\n" + msg.value;
        } else if (msg.type == "name") {
            document.getElementById("log").textContent += "\n" + msg.value;
        }
        
    } catch {

    }
};

socket.onerror = error => {
    console.log("Socket Error: ", error);
};

function sendChat() {
    
    //conn.send(document.getElementById("input").value);
    let inputValue = document.getElementById("input").value;
    let jmsg = JSON.stringify({
      type: "chat",
      value: inputValue
    });
    console.log("jmsg " + jmsg);
    socket.send(jmsg);
  
    // let pmsg = JSON.stringify({
    //   type: "ping", 
    //   value: "ping"   
    // });
  
    // conn.send(pmsg);
    //conn.send("ping");
  }

  function registerName() {
    
    //conn.send(document.getElementById("input").value);
    let inputValue = document.getElementById("registerInput").value;
    console.log("inputValue " + inputValue);
    let jmsg = JSON.stringify({
      type: "name",
      value: inputValue
    });
    console.log("jmsg " + jmsg);
    socket.send(jmsg);
  
  
  }
  
  document.getElementById("btn").addEventListener("click", sendChat);
  document.getElementById("registerButton").addEventListener("click", registerName);