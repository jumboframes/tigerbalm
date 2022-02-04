var http = require("http")
var log = require("log")

function register() {
    var route = new Object();
    route.match = new Object();
    route.match.path = "/foo";
    route.match.method = "GET";
    route.handler = httpHandler;
    var registration = Object();
    registration.route = route;
    return registration
}

function httpHandler(request) {
    //log.Debugf("incoming request, host: %s, url: %s, method: %s",
    //    request["Host"], request["Url"], request["Method"])
    log.Debugf("incoming request, host: %s, url: %s",
        request["Host"], request["Url"])

    req = {
        "Method": "GET",
        "Host": "127.0.0.1:1202",
        "Path": "/bar",
    }
    data = http.DoRequest(req)
    return {
        "Status": 200,
        "Body": "Austin Zhai"
    }
}
