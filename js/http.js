var http = require("http")
var log = require("log")
var env = require("env")

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
    log.Debugf("incoming request, host: %s, url: %s",
        request["Host"], request["Url"])

    req = {
        "Method": "GET",
        "Host": env.Get("BAIDU"),
        "Path": "/bar",
    }
    data = http.DoRequest(req)
    log.Debugf("do request, host: %s, body: %s", env.Get("BAIDU"), data["Body"])
    return {
        "Status": 200,
        "Body": "Austin Zhai"
    }
}
