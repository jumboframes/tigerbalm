# TigerBalm

Jack of all trades, A Faas(function as a service) framework, to custom a plugin by adding a javascript snippet, to activate the snippet by killing pid with signal hangup(```kill -HUP $pid```) instead of further compiling.

Currently support:

* env
* log
* http
* kafka

## To run tigerbalm

```
git clone https://github.com/jumboframes/tigerbalm
make
./tigerbalm -f ./tigerbalm.yaml

```

## To start customizing a snippet

### A proxy

```
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
        "Host": env.Get("GOOGLE"),
        "Path": "/bar",
    }
    data = http.DoRequest(req)
    log.Debugf("do request, host: %s, body: %s", env.Get("GOOGLE"), data["Body"])
    return {
        "Status": 200,
        "Body": "Austin Zhai"
    }
}

```

### A kafka logic broker

```
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

```