var log = require("log")
var producer = require("producer")

function register() {
    registration = {
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
    log.Debugf("incoming message, topic: %s, group: %s",
        msg["Topic"], msg["Group"])
    msg = {
        "Topic": "austin_relay",
        "Payload": msg["Payload"],
    }
    ret = producer.Produce(msg)
    log.Debugf("relay message to topic: %s %v", "austin_relay", ret)
}
