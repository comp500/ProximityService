const socket = new WebSocket(((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + "/ws");

// Connection opened
socket.addEventListener('open', function (event) {
});

// Listen for messages
socket.addEventListener('message', function (event) {
	console.log('Message from server ', event.data);
});