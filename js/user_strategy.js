function register() {
    var route = new Object();
    route.match = new Object();
    route.match.url = "/v1/users/strategy";
    route.match.method = "PUT";
    route.handler = handler;
    return route
}

function handler(request) {
    header = request["Header"]
    data = doRequest("GET", "192.168.111.13:4002", "/v1/user", header)
    body = JSON.parse(data["Body"])
    body = JSON.parse(data["Body"])
    if (body["code"] != 200) {
        return newResponse(data["Status"], data["Body"])
    }
    bodyData = body["data"]
    ids = []
    groups = bodyData["groups"]
    for (i = 0; i < groups.length; i++) {
        groups.push(groups[i]["id"])
    }
    ancestorGroups = bodyData["ancestor_groups"]
}

function newResponse(status, body) {
    var rsp = new Object();
    rsp.status = status
    rsp.body = body;
    return rsp;
}