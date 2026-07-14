let socket;

function connectWebSocket(){

    socket = new WebSocket("ws://localhost:8080/ws");

    socket.onopen = () => {
        console.log("Connected to Go Server");
    };

    socket.onclose = () => {
        console.log("Disconnected");
    };

    socket.onerror = (err) => {
        console.error(err);
    };

    socket.onmessage = (event) => {

        const candle = JSON.parse(event.data);

        console.log(candle);
        console.log(event.data);

        addCandle(candle);

    };

}