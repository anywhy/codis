'use strict';

var dashboard = angular.module('dashboard-fe', ["highcharts-ng", "ui.bootstrap"]);

function genXAuth(name) {
    return sha256("Codis-XAuth-[" + name + "]").substring(0, 32);
}

function concatUrl(base, name) {
    if (name) {
        return encodeURI(base + "?forward=" + name);
    } else {
        return encodeURI(base);
    }
}

function padInt2Str(num, size) {
    var s = num + "";
    while (s.length < size) s = "0" + s;
    return s;
}

function toJsonHtml(obj) {
    var json = angular.toJson(obj, 4);
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    json = json.replace(/ /g, '&nbsp;');
    json = json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        var cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
    return json;
}

function humanSize(size) {
    if (size < 1024) {
        return size + " B";
    }
    size /= 1024;
    if (size < 1024) {
        return size.toFixed(3) + " KB";
    }
    size /= 1024;
    if (size < 1024) {
        return size.toFixed(3) + " MB";
    }
    size /= 1024;
    if (size < 1024) {
        return size.toFixed(3) + " GB";
    }
    size /= 1024;
    return size.toFixed(3) + " TB";
}

function newChatsOpsConfig() {
    return {
        options: {
            chart: {
                useUTC: false,
                type: 'spline',
            },
        },
        series: [{
            color: '#d82b28',
            lineWidth: 1.5,
            states: {
                hover: {
                    enabled: false,
                }
            },
            showInLegend: false,
            marker: {
                enabled: true,
                symbol: 'circle',
                radius: 2,
            },
            name: 'OP/s',
            data: [],
        }],
        title: {
            style: {
                display: 'none',
            }
        },
        xAxis: {
            type: 'datetime',
            title: {
                style: {
                    display: 'none',
                }
            },
            labels: {
                formatter: function () {
                    var d = new Date(this.value);
                    return padInt2Str(d.getHours(), 2) + ":" + padInt2Str(d.getMinutes(), 2) + ":" + padInt2Str(d.getSeconds(), 2);
                }
            },
        },
        yAxis: {
            min: 0,
            title: {
                style: {
                    display: 'none',
                }
            },
        },
    };
}

function renderSlotsCharts(slots_array) {
    var groups = {};
    var counts = {};
    var n = slots_array.length;
    for (var i = 0; i < n; i++) {
        var slot = slots_array[i];
        groups[slot.group_id] = true;
        if (slot.action.target_id) {
            groups[slot.action.target_id] = true;
        }
        if (counts[slot.group_id]) {
            counts[slot.group_id]++;
        } else {
            counts[slot.group_id] = 1;
        }
    }
    var series = [];
    for (var g in groups) {
        var xaxis = 2;
        if (g == 0) {
            xaxis = 0;
        }
        var s = {name: 'Group-' + g + ':' + (counts[g] == undefined ? 0 : counts[g]), data: [], group_id: g};
        for (var beg = 0, end = 0; end <= n; end++) {
            if (end == n || slots_array[end].group_id != g) {
                if (beg < end) {
                    s.data.push({x: xaxis, low: beg, high: end - 1, group_id: g});
                }
                beg = end + 1;
            }
        }
        xaxis = 1;
        for (var beg = 0, end = 0; end <= n; end++) {
            if (end == n || !(slots_array[end].action.target_id && slots_array[end].action.target_id == g)) {
                if (beg < end) {
                    s.data.push({x: xaxis, low: beg, high: end - 1, group_id: g});
                }
                beg = end + 1;
            }
        }
        s.data.sort(function (a, b) {
            return a.x - b.x;
        });
        series.push(s);
    }
    series.sort(function (a, b) {
        return a.group_id - b.group_id;
    });
    new Highcharts.Chart({
        chart: {
            renderTo: 'slots_charts',
            type: 'columnrange',
            inverted: true,
        },
        title: {
            style: {
                display: 'none',
            }
        },
        xAxis: {
            categories: ['Offline', 'Migrating', 'Default'],
            min: 0,
            max: 2,
        },
        yAxis: {
            min: 0,
            max: 1024,
            tickInterval: 64,
            title: {
                style: {
                    display: 'none',
                }
            },
        },
        legend: {
            enabled: true
        },
        plotOptions: {
            columnrange: {
                grouping: false
            },
            series: {
                animation: false,
                events: {
                    legendItemClick: function () {
                        return false;
                    },
                }
            },
        },
        credits: {
            enabled: false
        },
        tooltip: {
            formatter: function () {
                switch (this.point.x) {
                case 0:
                    return '<b>Slot-[' + this.point.low + "," + this.point.high + "]</b> are <b>Offline</b>";
                case 1:
                    return '<b>Slot-[' + this.point.low + "," + this.point.high + "]</b> will be moved to <b>Group-[" + this.point.group_id + "]</b>";
                case 2:
                    return '<b>Slot-[' + this.point.low + "," + this.point.high + "]</b> --> <b>Group-[" + this.point.group_id + "]</b>";
                }
            }
        },
        series: series,
    });
}

function processProxyStats(codis_stats) {
    var proxy_array = codis_stats.proxy.models;
    var proxy_stats = codis_stats.proxy.stats;
    var qps = 0, sessions = 0;
    for (var i = 0; i < proxy_array.length; i++) {
        var p = proxy_array[i];
        var s = proxy_stats[p.token];
        p.sessions = "NA";
        p.commands = "NA";
        if (!s) {
            p.status = "PENDING";
        } else if (s.timeout) {
            p.status = "TIMEOUT";
        } else if (s.error) {
            p.status = "ERROR";
        } else {
            if (s.stats.online) {
                p.sessions = "total=" + s.stats.sessions.total + ",alive=" + s.stats.sessions.alive;
                p.commands = "total=" + s.stats.ops.total + ",fails=" + s.stats.ops.fails + ",qps=" + s.stats.ops.qps;
                p.status = "HEALTHY";
            } else {
                p.status = "PENDING";
            }
            qps += s.stats.ops.qps;
            sessions += s.stats.sessions.alive;
        }
    }
    return {proxy_array: proxy_array, qps: qps, sessions: sessions};
}

function alertAction(text, callback) {
    BootstrapDialog.show({
        title: "Warning !!",
        message: text,
        closable: true,
        buttons: [{
            label: "OK",
            cssClass: "btn-primary",
            action: function (dialog) {
                dialog.close();
                callback();
            },
        }],
    });
}

function alertErrorResp(failedResp) {
    var text = "error response";
    if (failedResp.status != 1500 && failedResp.status != 800) {
        text = failedResp.data.toString();
    } else {
        text = toJsonHtml(failedResp.data);
    }
    BootstrapDialog.alert({
        title: "Error !!",
        type: "type-danger",
        closable: true,
        message: text,
    });
}

function isValidInput(text) {
    return text && text != "" && text != "NA";
}

function processGroupStats(codis_stats) {
    var group_array = codis_stats.group.models;
    var group_stats = codis_stats.group.stats;
    var keys = 0, memory = 0;
    for (var i = 0; i < group_array.length; i++) {
        var g = group_array[i];
        if (g.promoting.state) {
            g.ispromoting = g.promoting.state != "";
        } else {
            g.ispromoting = false;
        }
        g.canremove = (g.servers.length == 0);
        for (var j = 0; j < g.servers.length; j++) {
            var x = g.servers[j];
            var s = group_stats[x.server];
            x.keys = "NA";
            x.memory = "NA";
            x.maxmem = "NA";
            x.master = "NA";
            if (!s) {
                x.status = "PENDING";
            } else if (s.timeout) {
                x.status = "TIMEOUT";
            } else if (s.error) {
                x.status = "ERROR";
            } else {
                if (s.stats["db0"]) {
                    var v = parseInt(s.stats["db0"].split(",")[0].split("=")[1], 10);
                    if (j == 0) {
                        keys += v;
                    }
                    x.keys = s.stats["db0"];
                }
                if (s.stats["used_memory"]) {
                    var v = parseInt(s.stats["used_memory"], 10);
                    if (j == 0) {
                        memory += v;
                    }
                    x.memory = humanSize(v);
                }
                if (s.stats["maxmemory"]) {
                    var v = parseInt(s.stats["maxmemory"], 10);
                    if (v == 0) {
                        x.maxmem = "INF."
                    } else {
                        x.maxmem = humanSize(v);
                    }
                }
                if (s.stats["master_addr"]) {
                    x.master = s.stats["master_addr"] + ":" + s.stats["master_link_status"];
                } else {
                    x.master = "NO:ONE";
                }
                if (j == 0) {
                    x.master_status = (x.master == "NO:ONE");
                } else {
                    x.master_status = (x.master == g.servers[0].server + ":up");
                }
            }
            if (g.ispromoting) {
                x.canremove = false;
                x.canpromote = false;
                x.canslaveof = "";
                x.actionstate = "";
            } else {
                x.canremove = (j != 0 || g.servers.length <= 1);
                x.canpromote = j != 0;
                if (x.action.state) {
                    if (x.action.state != "pending") {
                        x.canslaveof = "create";
                        x.actionstate = x.action.state;
                    } else {
                        x.canslaveof = "remove";
                        x.actionstate = x.action.state + ":" + x.action.index;
                    }
                } else {
                    x.canslaveof = "create";
                    x.actionstate = "";
                }
            }
        }
    }
    return {group_array: group_array, keys: keys, memory: memory};
}

dashboard.config(['$interpolateProvider',
    function ($interpolateProvider) {
        $interpolateProvider.startSymbol('[[');
        $interpolateProvider.endSymbol(']]');
    }
]);

dashboard.config(['$httpProvider', function ($httpProvider) {
    $httpProvider.defaults.useXDomain = true;
    delete $httpProvider.defaults.headers.common['X-Requested-With'];
}]);

dashboard.controller('MainCodisCtrl', ['$scope', '$http', '$uibModal', '$timeout',
    function ($scope, $http, $uibModal, $timeout) {
        Highcharts.setOptions({
            global: {
                useUTC: false,
            },
            exporting: {
                enabled: false,
            },
        });
        $scope.chart_ops = newChatsOpsConfig();

        $scope.refresh_interval = 3;

        $scope.resetOverview = function () {
            $scope.codis_name = "NA";
            $scope.codis_addr = "NA";
            $scope.codis_coord = "NA";
            $scope.codis_qps = "NA";
            $scope.codis_sessions = "NA";
            $scope.redis_mem = "NA";
            $scope.redis_keys = "NA";
            $scope.slots_array = [];
            $scope.proxy_array = [];
            $scope.group_array = [];
            $scope.slots_actions = [];
            $scope.chart_ops.series[0].data = [];
            $scope.slots_action_interval = "NA";
            $scope.slots_action_disabled = "NA";
            $scope.slots_action_failed = false;
            $scope.slots_action_remain = 0;
        }
        $scope.resetOverview();

        $http.get('/list').then(function (resp) {
            $scope.codis_list = resp.data;
        });

        $scope.selectCodisInstance = function (selected) {
            if ($scope.codis_name == selected) {
                return;
            }
            $scope.resetOverview();
            $scope.codis_name = selected;
            var url = concatUrl("/topom", selected);
            $http.get(url).then(function (resp) {
                if ($scope.codis_name != selected) {
                    return;
                }
                var overview = resp.data;
                $scope.codis_addr = overview.model.admin_addr;
                $scope.codis_coord = "[" + overview.config.coordinator_name + "] " + overview.config.coordinator_addr;
                $scope.updateStats(overview.stats);
            });
        }

        $scope.updateStats = function (codis_stats) {
            var proxy_stats = processProxyStats(codis_stats);
            var group_stats = processGroupStats(codis_stats);

            $scope.codis_qps = proxy_stats.qps;
            $scope.codis_sessions = proxy_stats.sessions;
            $scope.redis_mem = humanSize(group_stats.memory);
            $scope.redis_keys = group_stats.keys.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
            $scope.slots_array = codis_stats.slots;
            $scope.proxy_array = proxy_stats.proxy_array;
            $scope.group_array = group_stats.group_array;
            $scope.slots_actions = [];
            $scope.slots_action_interval = codis_stats.slot_action.interval;
            $scope.slots_action_disabled = codis_stats.slot_action.disabled;
            $scope.slots_action_failed = codis_stats.slot_action.progress.failed;
            $scope.slots_action_remain = codis_stats.slot_action.progress.remain;

            for (var i = 0; i < $scope.slots_array.length; i++) {
                var slot = $scope.slots_array[i];
                if (slot.action.state) {
                    $scope.slots_actions.push(slot);
                }
            }

            renderSlotsCharts($scope.slots_array);

            var ops_array = $scope.chart_ops.series[0].data;
            if (ops_array.length >= 10) {
                ops_array.shift();
            }
            ops_array.push({x: new Date(), y: proxy_stats.qps});
            $scope.chart_ops.series[0].data = ops_array;
        }

        $scope.refreshStats = function () {
            var codis_name = $scope.codis_name;
            var codis_addr = $scope.codis_addr;
            if (isValidInput(codis_name) && isValidInput(codis_addr)) {
                var xauth = genXAuth(codis_name);
                var url = concatUrl("/api/topom/stats/" + xauth, codis_name);
                $http.get(url).then(function (resp) {
                    if ($scope.codis_name != codis_name) {
                        return;
                    }
                    $scope.updateStats(resp.data);
                });
            }
        }

        $scope.createProxy = function (proxy_addr) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name) && isValidInput(proxy_addr)) {
                var xauth = genXAuth(codis_name);
                var url = concatUrl("/api/topom/proxy/create/" + xauth + "/" + proxy_addr, codis_name);
                $http.put(url).then(function () {
                    $scope.refreshStats();
                }, function (failedResp) {
                    alertErrorResp(failedResp);
                });
            }
        }

        $scope.removeProxy = function (proxy) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name)) {
                alertAction("Remove and Shutdown proxy: " + toJsonHtml(proxy), function () {
                    var xauth = genXAuth(codis_name);
                    var url = concatUrl("/api/topom/proxy/remove/" + xauth + "/" + proxy.token + "/0", codis_name);
                    $http.put(url).then(function () {
                        $scope.refreshStats();
                    }, function (failedResp) {
                        alertErrorResp(failedResp);
                    });
                });
            }
        }

        $scope.reinitProxy = function (proxy) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name)) {
                alertAction("Reinit and Start proxy: " + toJsonHtml(proxy), function () {
                    var xauth = genXAuth(codis_name);
                    var url = concatUrl("/api/topom/proxy/reinit/" + xauth + "/" + proxy.token, codis_name);
                    $http.put(url).then(function () {
                        $scope.refreshStats();
                    }, function (failedResp) {
                        alertErrorResp(failedResp);
                    });
                });
            }
        }

        $scope.createGroup = function (group_id) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name) && isValidInput(group_id)) {
                var xauth = genXAuth(codis_name);
                var url = concatUrl("/api/topom/group/create/" + xauth + "/" + group_id, codis_name);
                $http.put(url).then(function () {
                    $scope.refreshStats();
                }, function (failedResp) {
                    alertErrorResp(failedResp);
                });
            }
        }

        $scope.removeGroup = function (group_id) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name)) {
                var xauth = genXAuth(codis_name);
                var url = concatUrl("/api/topom/group/remove/" + xauth + "/" + group_id, codis_name);
                $http.put(url).then(function () {
                    $scope.refreshStats();
                }, function (failedResp) {
                    alertErrorResp(failedResp);
                });
            }
        }

        $scope.addGroupServer = function (group_id, server_addr) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name) && isValidInput(group_id) && isValidInput(server_addr)) {
                var xauth = genXAuth(codis_name);
                var url = concatUrl("/api/topom/group/add/" + xauth + "/" + group_id + "/" + server_addr, codis_name);
                $http.put(url).then(function () {
                    $scope.refreshStats();
                }, function (failedResp) {
                    alertErrorResp(failedResp);
                });
            }
        }

        $scope.delGroupServer = function (group_id, server_addr) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name) && isValidInput(group_id) && isValidInput(server_addr)) {
                alertAction("Remove server " + server_addr + " from Group-[" + group_id + "]", function () {
                    var xauth = genXAuth(codis_name);
                    var url = concatUrl("/api/topom/group/del/" + xauth + "/" + group_id + "/" + server_addr, codis_name);
                    $http.put(url).then(function () {
                        $scope.refreshStats();
                    }, function (failedResp) {
                        alertErrorResp(failedResp);
                    });
                });
            }
        }

        $scope.promoteServer = function (group_id, server_addr) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name) && isValidInput(group_id) && isValidInput(server_addr)) {
                alertAction("Promote server " + server_addr + " from Group-[" + group_id + "]", function () {
                    var xauth = genXAuth(codis_name);
                    var url = concatUrl("/api/topom/group/promote/" + xauth + "/" + group_id + "/" + server_addr, codis_name);
                    $http.put(url).then(function () {
                        $scope.promoteCommit(group_id);
                    }, function (failedResp) {
                        alertErrorResp(failedResp);
                    });
                });
            }
        }

        $scope.promoteCommit = function (group_id) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name) && isValidInput(group_id)) {
                var xauth = genXAuth(codis_name);
                var url = concatUrl("/api/topom/group/promote-commit/" + xauth + "/" + group_id, codis_name);
                $http.put(url).then(function () {
                    $scope.refreshStats();
                }, function (failedResp) {
                    alertErrorResp(failedResp);
                });
            }
        }

        $scope.createSyncAction = function (server_addr) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name) && isValidInput(server_addr)) {
                var xauth = genXAuth(codis_name);
                var url = concatUrl("/api/topom/group/action/create/" + xauth + "/" + server_addr, codis_name);
                $http.put(url).then(function () {
                    $scope.refreshStats();
                }, function (failedResp) {
                    alertErrorResp(failedResp);
                });
            }
        }

        $scope.removeSyncAction = function (server_addr) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name) && isValidInput(server_addr)) {
                var xauth = genXAuth(codis_name);
                var url = concatUrl("/api/topom/group/action/remove/" + xauth + "/" + server_addr, codis_name);
                $http.put(url).then(function () {
                    $scope.refreshStats();
                }, function (failedResp) {
                    alertErrorResp(failedResp);
                });
            }
        }

        $scope.createSlotAction = function (slot_id, group_id) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name) && isValidInput(slot_id) && isValidInput(group_id)) {
                alertAction("Migrate Slots-[" + slot_id + "] to Group-[" + group_id + "]", function () {
                    var xauth = genXAuth(codis_name);
                    var url = concatUrl("/api/topom/slots/action/create/" + xauth + "/" + slot_id + "/" + group_id, codis_name);
                    $http.put(url).then(function () {
                        $scope.refreshStats();
                    }, function (failedResp) {
                        alertErrorResp(failedResp);
                    });
                });
            }
        }

        $scope.createSlotActionRange = function (slot_beg, slot_end, group_id) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name) && isValidInput(slot_beg) && isValidInput(slot_end) && isValidInput(group_id)) {
                alertAction("Migrate Slots-[" + slot_beg + "," + slot_end + "] to Group-[" + group_id + "]", function () {
                    var xauth = genXAuth(codis_name);
                    var url = concatUrl("/api/topom/slots/action/create-range/" + xauth + "/" + slot_beg + "/" + slot_end + "/" + group_id, codis_name);
                    $http.put(url).then(function () {
                        $scope.refreshStats();
                    }, function (failedResp) {
                        alertErrorResp(failedResp);
                    });
                });
            }
        }

        $scope.removeSlotAction = function (slot_id) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name) && isValidInput(slot_id)) {
                var xauth = genXAuth(codis_name);
                var url = concatUrl("/api/topom/slots/action/remove/" + xauth + "/" + slot_id, codis_name);
                $http.put(url).then(function () {
                    $scope.refreshStats();
                }, function (failedResp) {
                    alertErrorResp(failedResp);
                });
            }
        }

        $scope.updateSlotActionDisabled = function (value) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name)) {
                var xauth = genXAuth(codis_name);
                var url = concatUrl("/api/topom/slots/action/disabled/" + xauth + "/" + value, codis_name);
                $http.put(url).then(function () {
                    $scope.refreshStats();
                }, function (failedResp) {
                    alertErrorResp(failedResp);
                });
            }
        }

        $scope.updateSlotActionInterval = function (value) {
            var codis_name = $scope.codis_name;
            if (isValidInput(codis_name)) {
                var xauth = genXAuth(codis_name);
                var url = concatUrl("/api/topom/slots/action/interval/" + xauth + "/" + value, codis_name);
                $http.put(url).then(function () {
                    $scope.refreshStats();
                }, function (failedResp) {
                    alertErrorResp(failedResp);
                });
            }
        }

        if (window.location.hash) {
            $scope.selectCodisInstance(window.location.hash.substring(1));
        }

        var ticker = 0;
        (function autoRefreshStats() {
            if (ticker >= $scope.refresh_interval) {
                ticker = 0;
                $scope.refreshStats();
            }
            ticker++;
            $timeout(autoRefreshStats, 1000);
        }());
    }
])
;
