function register() {
    registration = {
        "route": {
            "match": {
                "path": "/foo",
                "method": "GET"
            },
            "handler": httpHandler,
        },
        "consume": {
            "match": {
                "topic": "austin",
                "group": "zhai"
            },
            "handler": kafkaHandler,
        }
    }
    return registration
}

function kafkaHandler(msg) {
}
