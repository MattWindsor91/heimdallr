var ws = new WebSocket("ws://" + location.host + "/ws")

ws.onmessage = function (event) {
    console.log(event.data);
    $("#events").append("<p>" + event.data + "</p>");
    $("#events").scrollTop($("#events")[0].scrollHeight);
}
