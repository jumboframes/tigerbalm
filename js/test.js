function register() {
    var route = new Object();
    route.match = new Object();
    route.match.url = "/foo";
    route.match.method = "GET";
    route.handler = handler;
    return route
}

function handler(request) {
    data = doRequest("192.168.180.55", "/", "GET", "")
    console.log(data)
    return newResponse(data)
}

function newResponse(body) {
    var rsp = new Object();
    rsp.body = body;
    return rsp;
}
