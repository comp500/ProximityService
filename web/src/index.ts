const socket = new WebSocket(
//	(window.location.protocol === "https:" ? "wss://" : "ws://") + window.location.host + "/ws"
	"ws://192.168.1.109:9000/ws"
);

// Connection opened
socket.addEventListener("open", function(event) {
	console.log("Connected");
});

class eventMessage {
	Bin: boolean;
	Analog: number;
}

let startTime = new Date().getTime();
let numMsgs = 0;
const digital = document.getElementById("digital");
const analog = document.getElementById("analog");
const dt = document.getElementById("dt");
const dps = document.getElementById("dps");
const total = document.getElementById("total");

// Listen for messages
socket.addEventListener("message", function(event) {
	//console.log("Message from server ", event.data);
	let msg = <eventMessage>JSON.parse(event.data);
	digital.innerText = msg.Bin.toString();
	analog.innerText = msg.Analog.toString();
	let newTime = new Date().getTime();
	//timeAvg = (timeAvg + (newTime - lastTime)) / 2;
	//dt.innerText = timeAvg.toString();
	//lastTime = newTime;
	numMsgs++;
	dt.innerText = ((newTime - startTime) / numMsgs).toString();
	dps.innerText = (numMsgs / ((newTime - startTime) / 1000)).toString();
	total.innerText = numMsgs.toString();
});
