function el(id) {
    return document.getElementById(id);
}
let angryColors = true;
function onLoad() {
    var _a, _b;
    const url = new URL(window.location.toString());
    const angryColorsElm = el('waterfall-angry-colors');
    angryColors = url.searchParams.get('angry-colors') ? true : false;
    if (angryColors) {
        angryColorsElm.checked = true;
    }
    angryColorsElm.addEventListener('change', (e) => {
        if (e.target.checked) {
            angryColors = true;
            setLocationQueryParam('angry-colors', 'true');
        }
        else {
            angryColors = false;
            clearLocationQueryParam('angry-colors');
        }
        if (currentTree) {
            emptyTimingsTable();
            renderTree(currentTree, 0);
        }
    });
    const urlElm = el('waterfall-request-url');
    if (url.searchParams.get('url')) {
        urlElm.value = url.searchParams.get('url');
    }
    (_a = el('waterfall-button-fetch')) === null || _a === void 0 ? void 0 : _a.addEventListener('click', () => {
        waterfallRequestSubmit();
    });
    (_b = el('waterfall-request-method')) === null || _b === void 0 ? void 0 : _b.addEventListener('change', (e) => {
        const selectElm = e.target;
        if (selectElm.value == "GET") {
            el('waterfall-body-holder').style.display = 'none';
        }
        else {
            el('waterfall-body-holder').style.display = 'block';
        }
    });
}
function setLocationQueryParam(param, value) {
    let url = new URL(document.location.toString());
    url.searchParams.set(param, value);
    window.history.replaceState({ path: url.toString() }, '', url.toString());
}
function clearLocationQueryParam(param) {
    let url = new URL(document.location.toString());
    url.searchParams.delete(param);
    window.history.replaceState({ path: url.toString() }, '', url.toString());
}
function waterfallRequestSubmit() {
    var _a, _b, _c;
    const method = (_a = el('waterfall-request-method')) === null || _a === void 0 ? void 0 : _a.value;
    const path = (_b = el('waterfall-request-url')) === null || _b === void 0 ? void 0 : _b.value;
    const bodyTxt = (_c = el('waterfall-request-body')) === null || _c === void 0 ? void 0 : _c.value;
    let contentType = el('waterfall-body-type').value;
    setLocationQueryParam("url", path);
    makeWaterfallRequest(method, path, bodyTxt, contentType);
}
function setStatusText(str) {
    const statusElm = el('status-text');
    if (statusElm)
        statusElm.innerText = str;
}
let currentTree;
let abortFetch;
function makeWaterfallRequest(method, path, body, type) {
    let init = {
        method: method,
    };
    if (abortFetch != undefined) {
        // Abort request in progress
        abortFetch.abort();
    }
    if (abortFetch == undefined) {
        try {
            abortFetch = new AbortController();
            init.signal = abortFetch.signal;
        }
        catch (_a) {
            /* discard */
        }
    }
    if (type !== undefined) {
        init.headers = { 'Content-Type': type };
    }
    if (method != "GET" && method != "HEAD" && body !== undefined) {
        init.body = body;
    }
    emptyTimingsTable();
    currentTree = undefined;
    setStatusText('Making request...');
    fetch(path, init).then((r) => {
        abortFetch = undefined;
        if (r.status > 299 || r.status < 200) {
            setStatusText(`Server returned ${r.status}`);
            return;
        }
        const timingHeader = r.headers.get('Server-Timing');
        if (!timingHeader) {
            setStatusText('No Server-Timing headers in response');
            return;
        }
        setStatusText('');
        renderTimingsFromHeader(timingHeader);
    }).catch((e) => {
        abortFetch = undefined;
        if (e.name !== "AbortError") {
            setStatusText(`Failed to make request: ${e.message}`);
        }
    });
}
function splitHeader(header) {
    let inQuote = false;
    let start = 0;
    let timers = [];
    for (let i = 0; i < header.length; i++) {
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
    const timers = splitHeader(header);
    const re = new RegExp('([^;=]*)=("([^"]*)"|[^";]*)|([^=;]+);', 'g');
    let position = {};
    let startTime;
    let endTime;
    for (let timerStr of timers) {
        timerStr = timerStr + ';';
        let timer = { children: [] };
        let match;
        while ((match = re.exec(timerStr)) !== null) {
            let val;
            let name = match[1] ? match[1] : match[4];
            if (match[3] != undefined) {
                val = match[3];
            }
            else if (match[2] != undefined) {
                val = match[2];
            }
            else {
                val = match[4];
            }
            timer[name] = val;
        }
        let t = {
            id: timer['id'] !== undefined ? parseInt(timer['id']) : undefined,
            parent: timer['parent'] !== undefined ? parseInt(timer['parent']) : undefined,
            name: timer['descr'],
            start: parseInt(timer['start']),
            duration: parseFloat(timer['dur']),
            children: [],
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
    let root = [];
    let timer;
    for (timer of Object.values(position)) {
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
    for (timer of Object.values(position)) {
        if (timer.children.length > 0) {
            timer.children.sort((a, b) => {
                if (a.start != b.start)
                    return a.start - b.start;
                return a.dur - b.dur;
            });
        }
    }
    //console.log(root)
    let tree = {
        nodes: root,
        start: startTime,
        end: endTime,
    };
    return tree;
}
function renderTimingsFromHeader(header) {
    const tree = headerTimingToTree(header);
    currentTree = tree;
    emptyTimingsTable();
    renderTree(tree, 0);
}
function emptyTimingsTable() {
    const tBodyElm = el('waterfall-table-body');
    while (tBodyElm.firstChild)
        tBodyElm.removeChild(tBodyElm.firstChild);
}
function renderTree(tree, depth) {
    const tBodyElm = el('waterfall-table-body');
    for (const node of tree.nodes) {
        const rowElm = buildTableRow(node, depth, tree.start, tree.end);
        tBodyElm.appendChild(rowElm);
        renderTree({
            nodes: node.children,
            start: tree.start,
            end: tree.end
        }, depth + 1);
    }
}
function buildTableRow(node, depth, start, end) {
    const rowElm = document.createElement('tr');
    const nameCellElm = document.createElement('td');
    const timingElm = document.createElement('td');
    const barElm = document.createElement('div');
    const nameElm = document.createElement('span');
    nameCellElm.className = "waterfall-name-cell";
    timingElm.className = "waterfall-timer-cell";
    barElm.className = "waterfall-timer-bar";
    nameElm.className = "waterfall-timer-name";
    for (let i = 0; i < depth; i++) {
        const indentElm = document.createElement('span');
        indentElm.innerHTML = "&nbsp;";
        indentElm.className = "waterfall-indent";
        nameCellElm.appendChild(indentElm);
    }
    nameElm.innerText = node.name;
    nameCellElm.appendChild(nameElm);
    const totalDuration = end - start;
    const percentWidth = Math.round((node.duration / totalDuration) * 100);
    const percentOffset = Math.round(((node.start - start) / totalDuration) * 100);
    barElm.style.left = `${percentOffset}%`;
    barElm.style.width = `${percentWidth}%`;
    barElm.innerText = `${Math.round(node.duration * 10) / 10}ms`;
    if (angryColors) {
        barElm.style.backgroundImage = `linear-gradient(hsl(${100 - percentWidth}, 60%, 60%), hsl(${100 - percentWidth}, 60%, 40%))`;
    }
    timingElm.appendChild(barElm);
    rowElm.appendChild(nameCellElm);
    rowElm.appendChild(timingElm);
    return rowElm;
}
window.addEventListener('load', (event) => {
    onLoad();
});
