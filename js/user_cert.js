function register() {
    var route = new Object();
    route.match = new Object();
    route.match.url = "/v1/users/batch/cert_free";
    route.match.method = "PUT";
    route.handler = handler;
    return route
}

function handler(request) {
    // 获取
    data = JSON.parse(request["Body"])
    timestamp = data["timestamp"]
    userIds = data["user_ids"]

    // 聚合
    attributes = [];
    attributes[0] = {
        "k": "certification_free",
        "v": timestamp,
    }
    body = new Object();
    body["attributes"] = attributes;
    body["user_ids"] = userIds;
    
    str = JSON.stringify(body);
    data = doRequest("PUT", "192.168.111.13:4002", "/v1/users/batch", {}, str)
    return newResponse(data["Status"], data["Body"])
}

function newResponse(status, body) {
    var rsp = new Object();
    rsp.status = status
    rsp.body = body;
    return rsp;
}
