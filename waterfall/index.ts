
function el(id: string) : HTMLElement{
    return document.getElementById(id)
}

let angryColors = true

function onLoad() {
    const url = new URL(window.location.toString())

    const angryColorsElm = el('waterfall-angry-colors') as HTMLInputElement
    angryColors = url.searchParams.get('angry-colors') ? true : false
    if (angryColors) {
        angryColorsElm.checked = true
    }
    angryColorsElm.addEventListener('change', (e: Event) => {
        if ((e.target as HTMLInputElement).checked) {
            angryColors = true
            setLocationQueryParam('angry-colors', 'true')
        } else {
            angryColors = false
            clearLocationQueryParam('angry-colors')
        }
        if (currentTree) {
            emptyTimingsTable()
            renderTree(currentTree, 0)
        }
    })

    const urlElm = el('waterfall-request-url') as HTMLInputElement
    if (url.searchParams.get('url')) {
        urlElm.value = url.searchParams.get('url')
    }

    el('waterfall-button-fetch')?.addEventListener('click', () => {
        waterfallRequestSubmit()
    })
    el('waterfall-request-method')?.addEventListener('change', (e: Event) => {
        const selectElm = e.target as HTMLSelectElement
        if (selectElm.value == "GET") {
            el('waterfall-body-holder').style.display = 'none'
        } else {
            el('waterfall-body-holder').style.display = 'block'
        }
    })
}

function setLocationQueryParam(param:string, value:string) {
    let url = new URL(document.location.toString())
    url.searchParams.set(param, value)
    window.history.replaceState({path:url.toString()},'',url.toString());
}
function clearLocationQueryParam(param:string) {
    let url = new URL(document.location.toString())
    url.searchParams.delete(param)
    window.history.replaceState({path:url.toString()},'',url.toString());
}

function waterfallRequestSubmit() {
    const method = (el('waterfall-request-method') as HTMLSelectElement)?.value
    const path = (el('waterfall-request-url') as HTMLInputElement)?.value 
    const bodyTxt = (el('waterfall-request-body') as HTMLTextAreaElement)?.value 
    let contentType = (el('waterfall-body-type') as HTMLSelectElement).value
    setLocationQueryParam("url", path)
    makeWaterfallRequest(method, path, bodyTxt, contentType)
}

function setStatusText(str: string) {
    const statusElm = el('status-text')
    if (statusElm) statusElm.innerText = str
}

interface Timer {
    id?:number 
    name:string 
    start:number 
    duration:number
    parent?:number
    children:Array<Timer>
}
interface Tree {
    nodes:Array<Timer>
    start:number
    end:number
}

let currentTree:Tree 

function makeWaterfallRequest(method: string, path: string, body?:string, type?:string) {
    let init: RequestInit = {
        method: method,
        
    }
    if (type !== undefined) {
        init.headers = {'Content-Type': type}
    }
    if (method != "GET" && method != "HEAD" && body !== undefined) {
        init.body = body 
    }
    emptyTimingsTable()
    currentTree = undefined
    setStatusText('Making request...')
    fetch(path, init).then((r: Response) => {
        const timingHeader = r.headers.get('Server-Timing')
        if (!timingHeader) {
            setStatusText('No Server-Timing headers in response')
            return
        }
        setStatusText('')
        renderTimingsFromHeader(timingHeader)
    }).catch((e: any) => {
        setStatusText(`Failed to make request: ${e}`)
    })
} 
function splitHeader(header: string):Array<string> {
    let inQuote = false 
    let start = 0
    let timers: Array<string> = []
    for(let i = 0; i < header.length; i++) {
        if (header[i] == '"') {
            inQuote = inQuote ? false : true 
        } else if (header[i] == ',' && !inQuote) {
            timers.push(header.substring(start,i).trim())
            start = i+2
        }
    }
    return timers
}
function headerTimingToTree(header: string):Tree {
    const timers = splitHeader(header)
    const re = new RegExp('([^;=]*)=("([^"]*)"|[^";]*)|([^=;]+);', 'g')
    let position: {[key:number]:Timer} = {}
    let startTime:number;
    let endTime:number;
    for(let timerStr of timers) {
        timerStr = timerStr + ';'
        let timer = {children:[]}
        let match
        while ((match = re.exec(timerStr)) !== null) {
            let val
            let name = match[1] ? match[1] : match[4]
            if (match[3] != undefined) {
                val = match[3]
            } else if (match[2] != undefined) {
                val = match[2]
            } else {
                val = match[4]
            }
            timer[name] = val
        }
        let t:Timer = {
            id: timer['id'] !== undefined ? parseInt(timer['id']) : undefined,
            parent: timer['parent'] !== undefined ? parseInt(timer['parent']) : undefined,
            name: timer['descr'],
            start: parseInt(timer['start']),
            duration: parseFloat(timer['dur']),
            children: [],

        }
        if (t.id !== undefined) {
            position[t.id] = t
        }
        if (!startTime || t.start < startTime) {
            startTime = t.start
        }
        if (!endTime || t.start + t.duration > endTime) {
            endTime = t.start + t.duration
        }
    }

    let root:Array<Timer> = []
    let timer: any
    for(timer of Object.values(position)) {
        if (timer.id === undefined || timer.parent === undefined) {
            root.push(timer)
        } else if (position[timer.parent] === undefined) {
            root.push(timer)
        } else {
            position[timer.parent].children.push(timer)
        }
    }
    for(timer of Object.values(position)) {
        if (timer.children.length > 0) {
            timer.children.sort((a,b) => {
                if (a.start != b.start) return a.start - b.start 
                return a.dur - b.dur
            })
        }
    }
    //console.log(root)
    let tree:Tree = {
        nodes: root,
        start: startTime, 
        end: endTime,
    }
    return tree
}
function renderTimingsFromHeader(header: string) {
    const tree = headerTimingToTree(header)
    currentTree = tree
    emptyTimingsTable()
    renderTree(tree,0)
}
function emptyTimingsTable() {
    const tBodyElm = el('waterfall-table-body')
    while(tBodyElm.firstChild) tBodyElm.removeChild(tBodyElm.firstChild)
}


function renderTree(tree: Tree, depth: number) {
    const tBodyElm = el('waterfall-table-body')
    for(const node of tree.nodes) {
        const rowElm = buildTableRow(node, depth, tree.start, tree.end)
        tBodyElm.appendChild(rowElm)

        renderTree({
            nodes:node.children,
            start:tree.start,
            end:tree.end
        }, depth+1)
    }
}

function buildTableRow(node: Timer, depth:number, start:number, end:number): HTMLTableRowElement {
    const rowElm = document.createElement('tr') as HTMLTableRowElement
    const nameCellElm = document.createElement('td')
    const timingElm = document.createElement('td')
    const barElm = document.createElement('div')
    const nameElm = document.createElement('span')
    nameCellElm.className = "waterfall-name-cell"
    timingElm.className = "waterfall-timer-cell"
    barElm.className = "waterfall-timer-bar"
    nameElm.className = "waterfall-timer-name"

    for(let i = 0; i < depth; i++) {
        const indentElm = document.createElement('span')
        indentElm.innerHTML = "&nbsp;"
        indentElm.className = "waterfall-indent"
        nameCellElm.appendChild(indentElm)
    }
    nameElm.innerText = node.name
    nameCellElm.appendChild(nameElm)
    
    const totalDuration = end-start
    const percentWidth = Math.round((node.duration / totalDuration) * 100)
    const percentOffset = Math.round(((node.start - start) / totalDuration) * 100)
    barElm.style.left = `${percentOffset}%`
    barElm.style.width = `${percentWidth}%`
    barElm.innerText = `${Math.round(node.duration * 10)/10 }ms`

    if (angryColors) {
        barElm.style.backgroundImage = `linear-gradient(hsl(${100-percentWidth}, 60%, 60%), hsl(${100-percentWidth}, 60%, 40%))`
    }

    timingElm.appendChild(barElm)
    rowElm.appendChild(nameCellElm)
    rowElm.appendChild(timingElm)
    return rowElm
}

window.addEventListener('load', (event) => {
    onLoad()
})