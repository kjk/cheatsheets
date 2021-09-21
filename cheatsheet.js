function len(o) {
    if (o && o.length) {
        return o.length;
    }
    return 0;
}

function elById(id) {
    if (id[0] == "#") {
        id = id.substring(1);
    }
    return document.getElementById(id);
}

function getRandomInt(max) {
    return Math.floor(Math.random() * max);
}

function focusSearch() {
    const el = document.getElementById("cs-search-input");
    el.focus();
}

// create HTML to highlight part of s starting at idx and with length len
function hilightSearchResult(txt, matches) {
    var prevIdx = 0;
    var n = matches.length;
    var res = "";
    var s = "";
    // alternate non-higlighted and highlihted strings
    for (var i = 0; i < n; i++) {
        var el = matches[i];
        var idx = el[0];
        var len = el[1];

        var nonHilightLen = idx - prevIdx;
        if (nonHilightLen > 0) {
            s = txt.substring(prevIdx, prevIdx + nonHilightLen);
            res += `<span>${s}</span>`;
        }
        s = txt.substring(idx, idx + len);
        res += `<span class="hili">${s}</span>`;
        prevIdx = idx + len;
    }
    var txtLen = txt.length;
    nonHilightLen = txtLen - prevIdx;
    if (nonHilightLen > 0) {
        s = txt.substring(prevIdx, prevIdx + nonHilightLen);
        res += `<span>${s}</span>`;
    }
    return res;
}

const Escape = 27;
const Enter = 13;

function isEnter(ev) {
    return ev.which === Enter;
}

function isUp(ev) {
    return (ev.key == "ArrowUp") || (ev.key == "Up");
}

function isDown(ev) {
    return (ev.key == "ArrowDown") || (ev.key == "Down");
}

// navigation up is: Up or Ctrl-P
function isNavUp(ev) {
    if (isUp(ev)) {
        return true;
    }
    return ev.ctrlKey && (ev.keyCode === 80);
}

// navigation down is: Down or Ctrl-N
function isNavDown(ev) {
    if (isDown(ev)) {
        return true;
    }
    return ev.ctrlKey && (ev.keyCode === 78);
}

function dir(ev) {
    if (isNavUp(ev)) {
        return -1;
    }
    if (isNavDown(ev)) {
        return 1;
    }
    return 0;
}
