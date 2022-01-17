function register() {
    var route = new Object();
    route.match = new Object();
    route.match.url = "/v1/roles/{id}/permission";
    route.match.method = "GET";
    route.handler = handler;
    return route
}

var permissionList = [{
        "main_type": "用户管理",
        "main_tag": "userManage",
        "show": true,
        "permit": false,
        "sub_permission_list": [{
                "resource": {
                    "name": "用户总览",
                    "tag": "g:userOverview",
                    "permit": false
                },
                "resource_action_list": [{
                    "action": {
                        "name": "查看",
                        "tag": "g:get",
                        "permit": false
                    }
                }, ]
            },
            {
                "resource": {
                    "name": "用户列表",
                    "tag": "g:userList",
                    "permit": false
                },
                "resource_action_list": [{
                        "action": {
                            "name": "查看",
                            "tag": "g:get",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "用户启停用",
                            "tag": "g:enable",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "重置身份认证",
                            "tag": "g:resetUser",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "重制MFA",
                            "tag": "g:resetMFA",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "用户免认证",
                            "tag": "g:certificationFree",
                            "permit": false
                        }
                    }
                ]
            }
        ]
    },
    {
        "main_type": "连接管理",
        "main_tag": "connManage",
        "show": true,
        "permit": false,
        "sub_permission_list": [{
                "resource": {
                    "name": "连接器",
                    "tag": "g:conn",
                    "permit": false
                },
                "resource_action_list": [{
                        "action": {
                            "name": "查看",
                            "tag": "g:get",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "添加连接器",
                            "tag": "g:create",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "删除连接器",
                            "tag": "g:delete",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "编辑连接器",
                            "tag": "g:update",
                            "permit": false
                        }
                    }
                ]
            },
            {
                "resource": {
                    "name": "连接分组",
                    "tag": "g:connGroup",
                    "permit": false
                },
                "resource_action_list": [{
                        "action": {
                            "name": "查看",
                            "tag": "g:get",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "添加连接器组",
                            "tag": "g:create",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "删除连接器组",
                            "tag": "g:delete",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "编辑连接器组",
                            "tag": "g:update",
                            "permit": false
                        }
                    }
                ]
            }
        ]
    },
    {
        "main_type": "应用管理",
        "main_tag": "appManage",
        "show": true,
        "permit": false,
        "sub_permission_list": [{
            "resource": {
                "name": "应用列表",
                "tag": "g:resourceList",
                "permit": false
            },
            "resource_action_list": [{
                    "action": {
                        "name": "查看",
                        "tag": "g:get",
                        "permit": false
                    }
                },
                {
                    "action": {
                        "name": "添加应用",
                        "tag": "g:create",
                        "permit": false
                    }
                },
                {
                    "action": {
                        "name": "应用起停用",
                        "tag": "g:enable",
                        "permit": false
                    }
                },
                {
                    "action": {
                        "name": "删除应用",
                        "tag": "g:delete",
                        "permit": false
                    }
                },
                {
                    "action": {
                        "name": "编辑应用信息",
                        "tag": "g:update",
                        "permit": false
                    }
                },
                {
                    "action": {
                        "name": "编辑应用分配关系",
                        "tag": "g:updateStrategy",
                        "permit": false
                    }
                }
            ]
        }]
    },
    {
        "main_type": "角色管理",
        "main_tag": "roleManage",
        "show": false,
        "permit": false,
        "sub_permission_list": [{
            "resource": {
                "name": "角色管理",
                "tag": "g:roleManage",
                "permit": false
            },
            "resource_action_list": [{
                    "action": {
                        "name": "查看",
                        "tag": "g:get",
                        "permit": false
                    }
                },
                {
                    "action": {
                        "name": "添加角色",
                        "tag": "g:create",
                        "permit": false
                    }
                },
                {
                    "action": {
                        "name": "删除角色",
                        "tag": "g:delete",
                        "permit": false
                    }
                },
                {
                    "action": {
                        "name": "编辑角色名称",
                        "tag": "g:updateName",
                        "permit": false
                    }
                },
                {
                    "action": {
                        "name": "编辑关联用户",
                        "tag": "g:updateAssociatedUser",
                        "permit": false
                    }
                },
                {
                    "action": {
                        "name": "权限分配",
                        "tag": "g:permissionAssignment",
                        "permit": false
                    }
                }
            ]
        }]
    },
    {
        "main_type": "安全策略",
        "main_tag": "securityStrategy",
        "show": true,
        "permit": false,
        "sub_permission_list": [{
                "resource": {
                    "name": "MFA列表",
                    "tag": "g:MFAList",
                    "permit": false
                },
                "resource_action_list": [{
                        "action": {
                            "name": "查看",
                            "tag": "g:get",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "MFA起停用",
                            "tag": "g:enable",
                            "permit": false
                        }
                    }
                ]
            },
            {
                "resource": {
                    "name": "异常事件列表",
                    "tag": "g:abnormalEventList",
                    "permit": false,
                },
                "resource_action_list": [{
                        "action": {
                            "name": "查看",
                            "tag": "g:get",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "异常事件管理",
                            "tag": "g:manage",
                            "permit": false
                        }
                    }
                ]
            },
            {
                "resource": {
                    "name": "安全策略列表",
                    "tag": "g:securityStrategyList",
                    "permit": false
                },
                "resource_action_list": [{
                        "action": {
                            "name": "查看",
                            "tag": "g:get",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "安全策略管理",
                            "tag": "g:manage",
                            "permit": false
                        }
                    }
                ]
            }
        ]
    },
    {
        "main_type": "工单管理",
        "main_tag": "workOrderManage",
        "show": false,
        "permit": false,
        "sub_permission_list": [{
            "resource": {
                "name": "工单管理",
                "tag": "g:workOrderManage",
                "permit": false
            },
            "resource_action_list": [{
                    "action": {
                        "name": "查看",
                        "tag": "g:get",
                        "permit": false
                    }
                },
                {
                    "action": {
                        "name": "工单审批",
                        "tag": "g:approval",
                        "permit": false
                    }
                }
            ]
        }]
    },
    {
        "main_type": "日志审计",
        "main_tag": "audit",
        "show": true,
        "permit": false,
        "sub_permission_list": [{
                "resource": {
                    "name": "用户日志",
                    "tag": "g:userAudit",
                    "permit": false
                },
                "resource_action_list": [{
                    "action": {
                        "name": "查看",
                        "tag": "g:get",
                        "permit": false
                    }
                }]
            },
            {
                "resource": {
                    "name": "管理日志",
                    "tag": "g:manageAudit",
                    "permit": false
                },
                "resource_action_list": [{
                    "action": {
                        "name": "查看",
                        "tag": "g:get",
                        "permit": false
                    }
                }]
            }
        ]
    },
    {
        "main_type": "系统设置",
        "main_tag": "setting",
        "show": true,
        "permit": false,
        "sub_permission_list": [{
                "resource": {
                    "name": "认证配置",
                    "tag": "g:authSetting",
                    "permit": false
                },
                "resource_action_list": [{
                        "action": {
                            "name": "查看",
                            "tag": "g:get",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "编辑",
                            "tag": "g:update",
                            "permit": false
                        }
                    }
                ]
            },
            {
                "resource": {
                    "name": "日志留存",
                    "tag": "g:logRetention",
                    "permit": false
                },
                "resource_action_list": [{
                        "action": {
                            "name": "查看",
                            "tag": "g:get",
                            "permit": false
                        }
                    },
                    {
                        "action": {
                            "name": "编辑",
                            "tag": "g:update",
                            "permit": false
                        }
                    }
                ]
            }
        ]
    }
]

function handler(request) {
    // 获取
    url = request["Url"]
    header = request["Header"]
    data = doRequest("GET", "192.168.111.13:4002", url, header)
    body = JSON.parse(data["Body"])
    if (body["code"] != 200) {
        return newResponse(data["Status"], data["Body"])
    }
    bodyData = body["data"];
    var newPermissionList = clone(permissionList);
    for (i = 0; i < bodyData.length; i++) {
        // permissionList
        for (j = 0; j < newPermissionList.length; j++) {
            sub_permission_list = newPermissionList[j]["sub_permission_list"]
            for (k = 0; k < sub_permission_list.length; k++) {
                sub_permission = sub_permission_list[k]
                if (sub_permission["resource"]["tag"] == bodyData[i]["resource"]) {
                    newPermissionList[j]["permit"] = true
                    sub_permission["resource"]["permit"] = true
                    action_list = sub_permission["resource_action_list"]
                    for (l = 0; l < action_list.length; l++) {
                        action = action_list[l]
                        if (action["action"]["tag"] == bodyData[i]["action"]) {
                            action["action"]["permit"] = true
                        }
                    }
                }
            }
        }
    }

    for (i = newPermissionList.length - 1; i >= 0; i--) {
        if (newPermissionList[i]["permit"] == false) {
            newPermissionList.splice(i, 1)
            continue
        }
        sub_permission_list = newPermissionList[i]["sub_permission_list"]
        for (j = sub_permission_list.length - 1; j >= 0; j--) {
            sub_permission = sub_permission_list[j]
            if (sub_permission["resource"]["permit"] == false) {
                sub_permission_list.splice(j, 1)
                continue
            }
            action_list = sub_permission["resource_action_list"]
            for (k = action_list.length - 1; k >= 0; k--) {
                action = action_list[k]
                if (action["action"]["permit"] == false) {
                    action_list.splice(k, 1)
                }
            }
        }
    }
    ret = {}
    ret["code"] = 200
    ret["message"] = "OK"
    ret["data"] = newPermissionList
    return newResponse(200, JSON.stringify(ret))
}

function clone(obj) {
    return JSON.parse(JSON.stringify(obj));
}

function newResponse(status, body) {
    var rsp = new Object();
    rsp.status = status
    rsp.body = body;
    return rsp;
}

function iterate(obj) {
    for (var property in obj) {
        if (obj.hasOwnProperty(property)) {
            if (typeof obj[property] == "object") {
                iterate(obj[property]);
            } else {
                console.log(property + "   " + obj[property]);
            }
        }
    }
}