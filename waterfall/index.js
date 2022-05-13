function el(id) {
    return document.getElementById(id);
}
function onLoad() {
    var _a;
    (_a = el('waterfall-button-fetch')) === null || _a === void 0 ? void 0 : _a.addEventListener('click', function () {
        waterfallRequestSubmit();
    });
}
function waterfallRequestSubmit() {
    var _a, _b, _c, _d;
    var method = (_a = el('waterfall-request-method')) === null || _a === void 0 ? void 0 : _a.value;
    var path = (_b = el('waterfall-request-url')) === null || _b === void 0 ? void 0 : _b.value;
    var bodyTxt = (_c = el('waterfall-request-body')) === null || _c === void 0 ? void 0 : _c.value;
    var contentType = 'text/plain';
    switch ((_d = el('waterfall-body-type')) === null || _d === void 0 ? void 0 : _d.value) {
        case 'JSON':
            contentType = 'application/json';
            break;
        case 'XML':
            contentType = 'application/xml';
            break;
    }
    makeWaterfallRequest(method, path, bodyTxt, contentType);
}
function setStatusText(str) {
    var statusElm = el('status-text');
    if (statusElm)
        statusElm.innerText = str;
}
function makeWaterfallRequest(method, path, body, type) {
    var init = {
        method: method
    };
    if (type !== undefined) {
        init.headers = { 'Content-Type': type };
    }
    if (method != "GET" && method != "HEAD" && body !== undefined) {
        init.body = body;
    }
    console.log(path);
    fetch(path, init).then(function (r) {
        var timingHeader = r.headers.get('Server-Timing');
        if (!timingHeader) {
            setStatusText('Failed to get headers for response');
            return;
        }
        console.log(timingHeader);
        renderTimingsFromHeader(timingHeader);
    }); /*.catch((e: any) => {
        setStatusText(`Failed to make request: ${e}`)
    })*/
}
function splitHeader(header) {
    var inQuote = false;
    var start = 0;
    var timers = [];
    for (var i = 0; i < header.length; i++) {
        if (header[i] == '"') {
            inQuote = inQuote ? false : true;
        }
        else if (header[i] == ',' && !inQuote) {
            timers.push(header.substring(start, i).trim());
            start = i + 2;
        }
    }
    return timers;
}
function headerTimingToTree(header) {
    var timers = splitHeader(header);
    var re = new RegExp('([^;=]*)=("([^"]*)"|[^";]*)|([^=;]+);', 'g');
    var position = {};
    var startTime;
    var endTime;
    for (var _i = 0, timers_1 = timers; _i < timers_1.length; _i++) {
        var timerStr = timers_1[_i];
        timerStr = timerStr + ';';
        var timer_1 = { children: [] };
        var match = void 0;
        while ((match = re.exec(timerStr)) !== null) {
            var val = void 0;
            var name_1 = match[1] ? match[1] : match[4];
            if (match[3] != undefined) {
                val = match[3];
            }
            else if (match[2] != undefined) {
                val = match[2];
            }
            else {
                val = match[4];
            }
            timer_1[name_1] = val;
        }
        var t = {
            id: timer_1['id'] !== undefined ? parseInt(timer_1['id']) : undefined,
            parent: timer_1['parent'] !== undefined ? parseInt(timer_1['parent']) : undefined,
            name: timer_1['descr'],
            start: parseInt(timer_1['start']),
            duration: parseFloat(timer_1['dur']),
            children: []
        };
        if (t.id !== undefined) {
            position[t.id] = t;
        }
        if (!startTime || t.start < startTime) {
            startTime = t.start;
        }
        if (!endTime || t.start + t.duration > endTime) {
            endTime = t.start + t.duration;
        }
    }
    console.log(position);
    var root = [];
    var timer;
    for (var _a = 0, _b = Object.values(position); _a < _b.length; _a++) {
        timer = _b[_a];
        if (timer.id === undefined || timer.parent === undefined) {
            root.push(timer);
        }
        else if (position[timer.parent] === undefined) {
            root.push(timer);
        }
        else {
            position[timer.parent].children.push(timer);
        }
    }
    for (var _c = 0, _d = Object.values(position); _c < _d.length; _c++) {
        timer = _d[_c];
        if (timer.children.length > 0) {
            timer.children.sort(function (a, b) {
                if (a.start != b.start)
                    return a.start - b.start;
                return a.dur - b.dur;
            });
        }
    }
    console.log(root);
    var tree = {
        nodes: root,
        start: startTime,
        end: endTime
    };
    return tree;
}
function renderTimingsFromHeader(header) {
    var tree = headerTimingToTree(header);
    // TODO: Empty table body
    renderTree(tree, 0);
}
function renderTree(tree, depth) {
    var tBodyElm = el('waterfall-table-body');
    for (var _i = 0, _a = tree.nodes; _i < _a.length; _i++) {
        var node = _a[_i];
        var rowElm = buildTableRow(node, depth, tree.start, tree.end);
        tBodyElm.appendChild(rowElm);
        renderTree({
            nodes: node.children,
            start: tree.start,
            end: tree.end
        }, depth + 1);
    }
}
function buildTableRow(node, depth, start, end) {
    var rowElm = document.createElement('tr');
    var nameCellElm = document.createElement('td');
    var timingElm = document.createElement('td');
    var barElm = document.createElement('div');
    var nameElm = document.createElement('span');
    nameCellElm.className = "waterfall-name-cell";
    timingElm.className = "waterfall-timer-cell";
    barElm.className = "waterfall-timer-bar";
    nameElm.className = "waterfall-timer-name";
    for (var i = 0; i < depth; i++) {
        var indentElm = document.createElement('span');
        indentElm.innerHTML = "&nbsp;";
        indentElm.className = "waterfall-indent";
        nameCellElm.appendChild(indentElm);
    }
    nameElm.innerText = node.name;
    nameCellElm.appendChild(nameElm);
    var totalDuration = end - start;
    var percentWidth = Math.round((node.duration / totalDuration) * 100);
    var percentOffset = Math.round(((node.start - start) / totalDuration) * 100);
    barElm.style.left = "".concat(percentOffset, "%");
    barElm.style.width = "".concat(percentWidth, "%");
    barElm.innerText = "".concat(Math.round(node.duration * 10) / 10, "ms");
    timingElm.appendChild(barElm);
    rowElm.appendChild(nameCellElm);
    rowElm.appendChild(timingElm);
    return rowElm;
}
window.addEventListener('load', function (event) {
    onLoad();
});
