var domParser = new DOMParser();
var xmlDocumentCache = null;
var mousePressed = false;
var ctx = null;
var actionRunning = false;
var testAppPath = "";
var selectedNode = "";

var useNaturalTouches = true;
var isClick = true;
var startPositionX = 0;
var startPositionY = 0;

var screenScale = 1;

var sceneGraphWidth = 0;
var sceneGraphHeight = 0;
var screenshotWidth = 0;
var screenshotHeight = 0;
var canvasHeight = 0;
var canvasWidth = 0;
var canvasMaxWidth = 1024;
var canvasMaxHeight = 700;
var selectOffset = 10;


function Draw(x, y, isDown) {
    if (isDown) {
        ctx.beginPath();
        ctx.strokeStyle = '#000000';
        ctx.lineWidth = 2;
        ctx.lineJoin = "round";
        ctx.moveTo(lastX, lastY);
        ctx.lineTo(x, y);
        ctx.closePath();
        ctx.stroke();
    }
    lastX = x;
    lastY = y;
}

function DrawRect(x, y, width, height) {
    ctx.beginPath();
    ctx.strokeStyle = '#FF0000';
    ctx.lineWidth = 1;
    ctx.lineJoin = "round";
    ctx.rect(x, y, width, height);
    ctx.stroke();
}

function clearArea() {
    ctx.save()
    // Use the identity matrix while clearing the canvas
    ctx.setTransform(1, 0, 0, 1, 0, 0);
    ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);
    const image = document.getElementById('screenshot');
    ctx.scale(screenScale, screenScale);
    ctx.drawImage(image, 0, 0);
    ctx.restore()
}

function setCookie(cname, cvalue, exdays) {
    var d = new Date();
    d.setTime(d.getTime() + (exdays * 24 * 60 * 60 * 1000));
    var expires = "expires=" + d.toUTCString();
    document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/";
}

function getCookie(cname) {
    var name = cname + "=";
    var decodedCookie = decodeURIComponent(document.cookie);
    var ca = decodedCookie.split(';');
    for (var i = 0; i < ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0) == ' ') {
            c = c.substring(1);
        }
        if (c.indexOf(name) == 0) {
            return c.substring(name.length, c.length);
        }
    }
    return "";
}

function activity(show) {
    if (show) {
        $('#overlay').show();
        $('#activityIndicator').addClass("is-active");
    } else {
        $('#overlay').hide();
        $('#activityIndicator').removeClass("is-active");
    }
}

function notify(msg) {
    var notification = document.querySelector('.mdl-js-snackbar');
    notification.MaterialSnackbar.showSnackbar(
        {
            message: msg
        }
    )
}

function showStartAppDialog() {
    var dialog = document.getElementById('startAppDialog');
    dialogPolyfill.registerDialog(dialog);
    dialog.showModal();
}

function closeStartAppDialog() {
    var dialog = document.getElementById('startAppDialog');
    dialog.close();
}

function startAppDialogSessionStart() {
    sessionStart();
    closeStartAppDialog();
}

function updateState() {
    var sessionId = getCookie("session_id");
    if (sessionId !== "") {
        $('#startDevice').hide();
        $('#sessionStartButton').hide();
        $('#sessionStopButton').show();
        $('#refresh').show();
        $('#fileuploader').hide();
        $('#deviceSelection').hide();
        $('#appSelection').hide();
    } else {
        $('#startDevice').show();
        $('#sessionStartButton').show();
        $('#sessionStopButton').hide();
        $('#refresh').hide();
        $('#fileuploader').show();
        $('#deviceSelection').show();
        $('#appSelection').show();
    }
}

function getBaseURL() {
    var sessionId = getCookie("session_id");
    return "/wd/hub/session/" + sessionId + "/"
}

function sessionStart() {
    if (actionRunning) {
        return;
    }
    var deviceSelection = $('#deviceSelectionBoxValue');
    if (deviceSelection.val() === "") {
        return;
    }
    var appSelectBoxValue = $('#appSelectBoxValue');
    if (appSelectBoxValue.val() === "") {
        return;
    }

    actionRunning = true;
    activity(true);
    $.ajax({
        type: "POST",
        url: "/wd/hub/session",
        // The key needs to match your method's input parameter (case-sensitive).
        data: JSON.stringify({
            desiredCapabilities: {},
            requiredCapabilities: {
                device_id: deviceSelection.val(),
                app: appSelectBoxValue.val(),
            }
        }),
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            actionRunning = false;
            activity(false);
            setCookie("session_id", data.sessionId, 365);
            updateState();
            notify("device started");
            getScreenshot();
        },
        error: function (request, status, error) {
            actionRunning = false;
            activity(false);
            notify(request.responseJSON.message)
        },
        failure: function (errMsg) {
            actionRunning = false;
            activity(false);
            notify(errMsg);
        }
    });
}

function getTests() {
    if (actionRunning) {
        return;
    }
    actionRunning = true;
    activity(true);
    $.ajax({
        type: "POST",
        url: getBaseURL() + "execute",
        success: function (data) {
            actionRunning = false;
            activity(false);
            if (data !== "" && data.payload !== "") {
                console.log(data.payload)
            } else {
                notify("could not get tests")
            }
        },
        error: function (request, status, error) {
            actionRunning = false;
            activity(false);
            notify(request.responseJSON.message)
        },
        failure: function (errMsg) {
            actionRunning = false;
            activity(false);
            notify(errMsg)
        }
    });
}

function refresh() {
    getScreenshot();
}

function sessionStop() {
    if (actionRunning) {
        return;
    }
    actionRunning = true;
    activity(true);
    var sessionId = getCookie("session_id");
    if (sessionId !== "") {
        $.ajax({
            type: "DELETE",
            url: getBaseURL(),
            success: function (data) {
                actionRunning = false;
                activity(false);
                setCookie("session_id", "", 365);
                $("#screenshot").attr("src", "");
                $("#graph").prop('value', "");
                updateState();
            },
            error: function (request, status, error) {
                actionRunning = false;
                activity(false);
                notify(request.responseJSON.message)
            },
            failure: function (errMsg) {
                actionRunning = false;
                activity(false);
                notify(errMsg);
            }
        });
    }
}

function getScreenshot() {
    if (actionRunning) {
        return;
    }
    actionRunning = true;
    activity(true);
    $.ajax({
        type: "GET",
        url: getBaseURL() + "screen",
        success: function (data) {
            actionRunning = false;
            activity(false);
            if (data !== "" && data.value.data !== "" && data.payload !== "") {
                var document = atob(data.payload);
                $("#graph").prop('value', document);
                xmlDocumentCache = domParser.parseFromString(document, "text/xml");
                var rootNode = xmlDocumentCache.documentElement;

                sceneGraphWidth = rootNode.getAttribute("RectangleX");
                sceneGraphHeight = rootNode.getAttribute("RectangleY");

                screenshotWidth = data.value.width;
                screenshotHeight = data.value.height;

                screenScale = Math.min(canvasMaxWidth / screenshotWidth, canvasMaxHeight / screenshotHeight/*, 1*/)


                canvasHeight = screenshotHeight * screenScale;
                canvasWidth = screenshotWidth * screenScale;

                var screenImage = $("#screenshot")
                var canvas = $('#deviceCanvas');
                screenImage.attr("src", "data:image/png;base64," + data.value.data);
                canvas.attr("width", canvasWidth);
                canvas.attr("height", canvasHeight);

                clearArea();
                updateGraph();
                DrawRect(0,0,0,0);
            } else {
                notify("could not get screenshot")
            }
        },
        error: function (request, status, error) {
            actionRunning = false;
            clearArea();
            updateGraph();
            activity(false);
            notify(request.responseJSON.message)
        },
        failure: function (errMsg) {
            actionRunning = false;
            clearArea();
            updateGraph();
            activity(false);
            notify(errMsg)
        }
    });
}

function HighlightElement(node) {
    var x = (node.getAttribute("X") / sceneGraphWidth) * canvasWidth;
    var y = (node.getAttribute("Y") / sceneGraphHeight) * canvasHeight;
    var width = (node.getAttribute("RectangleX") / sceneGraphWidth) * canvasWidth;
    var height = (node.getAttribute("RectangleY") / sceneGraphHeight) * canvasHeight;
    DrawRect(x, y, width, height);
}

function visualXPath() {
    var value = $('#xpath').val();
    if (xmlDocumentCache == null) {
        return
    }
    clearArea();
    try {
        var nodes = xmlDocumentCache.evaluate(value, xmlDocumentCache, null, XPathResult.ANY_TYPE, null);
        var elements = "";
        elements += "<ul class=\"mdl-list\">";
        var node = null;
        while (node = nodes.iterateNext()) {
            if (node.nodeName === '#document') {
                continue;
            }
            var id = node.getAttribute("id");
            var cl = node.getAttribute("class");
            HighlightElement(node);
            elements += "<li class=\"mdl-list__item\"><span class=\"mdl-list__item-primary-content\" onclick='selectNode(" + id + ")' onmouseout='showElementInfos()' onmouseover='showElementInfos(" + id + ")' id=\"" + id + "\">" + getNodeName(node) + "</span></li>";
        }
        elements += "</ul>"
        $('#xpathSelection').html(elements);
    } catch (e) {

    }
}

var nextTouchActions = new Map();

function touchPosition(action, x, y) {
    if (actionRunning) {
        nextTouchActions.set(action, [x, y])
        return;
    }
    actionRunning = true;
    if (action === "up") {
        activity(true);
    }
    $.ajax({
        type: "POST",
        url: getBaseURL() + "touch/" + action,
        // The key needs to match your method's input parameter (case-sensitive).
        data: JSON.stringify({
            x: x,
            y: y
        }),
        contentType: "application/json; charset=utf-8",
        dataType: "json",
        success: function (data) {
            actionRunning = false;
            var next = nextTouchActions.get("move");
            if (next !== undefined && next.length === 2) {
                nextTouchActions.set("move", [])
                touchPosition("move", next[0], next[1])
                return;
            }
            next = nextTouchActions.get("up");
            if (next !== undefined && next.length === 2) {
                nextTouchActions.set("up", [])
                touchPosition("up", next[0], next[1])
                return;
            }
            if (action === "up" || action === "position") {
                activity(false);
                getScreenshot();
            }
        },
        error: function (request, status, error) {
            actionRunning = false;
            if (action === "up") {
                activity(false);
            }
            notify(request.responseJSON.message)
        },
        failure: function (errMsg) {
            actionRunning = false;
            if (action === "up") {
                activity(false);
            }
            notify(errMsg);
        }
    });
}

function buildPath(node) {
    if (node.parentNode) {
        return buildPath(node.parentNode) + "/" + node.nodeName
    }
    return "/" + node.nodeName
}

function copyXPath() {
    var copyText = document.getElementById("detailsXpath");
    if (copyText !== null) {
        copyText.select();
        copyText.setSelectionRange(0, 99999); /*For mobile devices*/
        document.execCommand("copy");
        // alert("Copied the text: " + copyText.value);
        notify("Copied xpaht: " + copyText.value)
    }
}

function showElementInfos(id) {
    var selectedID = selectedNode;
    if (id !== undefined) {
        selectedID = id;
    }
    var xpath = "//*[@id=" + selectedID + "]";
    var nodes = xmlDocumentCache.evaluate(xpath, xmlDocumentCache, null, XPathResult.ANY_TYPE, null);
    var result = nodes.iterateNext();
    if (result != null) {
        var html = "";
        var nodeName = result.nodeName;
        var index = 0;
        var numNodes = 0;
        if (result.parentNode != null) {
            for (var i = 0; i < result.parentNode.childNodes.length; i++) {
                if (result.parentNode.childNodes[i].nodeType === 3) {
                    continue;
                }
                var otherChild = result.parentNode.childNodes[i].nodeName;
                if (otherChild === nodeName) {
                    numNodes++;
                    if (result.parentNode.childNodes[i].getAttribute("id") == selectedID) {
                        index = numNodes;
                    }
                }
            }
        }
        var nodeXPath = buildPath(result);
        nodeXPath = nodeXPath.replace("/#document", "");
        if (numNodes > 1) {
            nodeXPath += "[" + index + "]";
        }
        for (var i = 0, atts = result.attributes, n = atts.length, arr = []; i < n; i++) {
            html += atts[i].nodeName + ": " + atts[i].nodeValue + "<br />";
        }
        html += "XPath:</br><div class=\"mdl-textfield mdl-js-textfield\"><input class=\"mdl-textfield__input\" type=\"text\" value=\"" + nodeXPath + "\" id=\"detailsXpath\"><label class=\"mdl-textfield__label\" for=\"detailsXpath\">" + "" + "</label></div>";
        // html +="<button class=\"mdl-button mdl-js-button mdl-button--primary\" onclick=\"copyXPath()\">Copy XPath</button>";
        $('#elementDetails').html(html);
        clearArea();
        HighlightElement(result);
    }
}

function locateAndHighlightElement(x, y, width, height) {
    if (xmlDocumentCache == null) {
        return;
    }
    var endX = x + width;
    var endY = y + height;
    var xpath = "//*[@X>" + x + "]" + "[@X<" + endX + "]" + "[@Y>" + y + "]" + "[@Y<" + endY + "]";
    var nodes = xmlDocumentCache.evaluate(xpath, xmlDocumentCache, null, XPathResult.ANY_TYPE, null);
    var result = nodes.iterateNext();
    clearArea();
    var elements = "<ul class=\"mdl-list\">";
    while (result) {
        var id = result.getAttribute("id");
        // var cl = result.getAttribute("class");
        HighlightElement(result);
        elements += "<li class=\"mdl-list__item\"><span class=\"mdl-list__item-primary-content\" onclick='selectNode(" + id + ")' onmouseout='showElementInfos()' onmouseover='showElementInfos(" + id + ")' id=\"" + id + "\">" + getNodeName(result) + "</span></li>";
        result = nodes.iterateNext();
    }
    elements += "</ul>"
    $('#xpathSelection').html(elements);
}

function hoverElement() {
}

function clickElement() {
}

function hoverElementPos(x, y) {
    clearArea();
    var xpath = "//*[@X<" + x + "]" + "[@RectangleX>" + x + "]" + "[@Y<" + y + "]" + "[@RectangleY>" + y + "]";
    var nodes = xmlDocumentCache.evaluate(xpath, xmlDocumentCache, null, XPathResult.ANY_TYPE, null);
    var result = nodes.iterateNext();
    while (result) {
        var id = result.getAttribute("id");
        HighlightElement(result);
        // var nodeXPath = buildPath(result);
        // nodeXPath = nodeXPath.replace("/#document", "");
        showElementInfos(id);
        result = nodes.iterateNext();
    }
}

function isControlMode() {
    return $('#controlMode').prop("checked")
}

function isNaturalTouch() {
    return useNaturalTouches;
}

function getCanvasPosX(e, canvas) {
    return e.pageX - canvas.offset().left
}

function getCanvasPosY(e, canvas) {
    return  e.pageY - canvas.offset().top
}

function keepAlive() {
    var sessionId = getCookie("session_id");
    if (sessionId !== "") {
        $.ajax({
            type: "GET",
            url: getBaseURL() + "ping",
            success: function (data) {
            },
            error: function (request, status, error) {
            },
            failure: function (errMsg) {
            }
        });
    
    }
}

var moveAction = null;
var upAction = null;

$(document).ready(function () {
    $.ajaxSetup({ cache: false });

    const canvas = document.getElementById('deviceCanvas');
    ctx = canvas.getContext('2d');
    const image = document.getElementById('screenshot');
    image.addEventListener('load', e => {
        clearArea();
    });
    clearArea();

    var deviceView = $('#deviceView');
    canvasMaxWidth = deviceView.width();
    canvasMaxHeight = deviceView.height();

    canvasHeight = canvasMaxHeight;
    canvasWidth = canvasMaxWidth;


    var deviceCanvas = $('#deviceCanvas');
    deviceCanvas.attr("width", canvasWidth);
    deviceCanvas.attr("height", canvasHeight);
    deviceCanvas.mousedown(function (e) {
        mousePressed = true;
        isClick = true;
        startPositionX = getCanvasPosX(e, $(this));
        startPositionY = getCanvasPosY(e, $(this));
        if (isControlMode() && isNaturalTouch()) {
            touchPosition("down", (startPositionX / canvasWidth) * sceneGraphWidth, (startPositionY / canvasHeight) * sceneGraphHeight);
            return;
        }
        DrawRect(startPositionX, startPositionY, 0, 0);
    })
    deviceCanvas.mousemove(function (e) {
        isClick = false;
        var posX = getCanvasPosX(e, $(this));
        var posY = getCanvasPosY(e, $(this));

        if (!mousePressed) {
            // hoverElement(posX, posY);
            return;
        }

        if (isControlMode() && isNaturalTouch()) {
            touchPosition("move", (posX / canvasWidth) * sceneGraphWidth, (posY / canvasHeight) * sceneGraphHeight);
            return;
        }

        var minX = Math.min(startPositionX, posX);
        var minY = Math.min(startPositionY, posY);
        var maxX = Math.max(startPositionX, posX);
        var maxY = Math.max(startPositionY, posY);

        clearArea()
        DrawRect(minX, minY, maxX - minX, maxY - minY);
    });
    deviceCanvas.mouseup(function (e) {
        mousePressed = false;
        var posX = getCanvasPosX(e, $(this));
        var posY = getCanvasPosY(e, $(this));

        if (isControlMode()) {
            if (isNaturalTouch()) {
                touchPosition("up", (posX / canvasWidth) * sceneGraphWidth, (posY / canvasHeight) * sceneGraphHeight);
            } else {
                touchPosition("position", (posX / canvasWidth) * screenshotWidth, (posY / canvasHeight) * screenshotHeight);
            }
            return;
        }

        clearArea()
        var minX = Math.min(startPositionX, posX) - selectOffset;
        var minY = Math.min(startPositionY, posY) - selectOffset;
        var maxX = Math.max(startPositionX, posX) + selectOffset;
        var maxY = Math.max(startPositionY, posY) + selectOffset;
        locateAndHighlightElement((minX / canvasWidth) * sceneGraphWidth, (minY / canvasHeight) * sceneGraphHeight, ((maxX - minX)/ canvasWidth) * sceneGraphWidth, ((maxY - minY)/ canvasHeight) * sceneGraphHeight)
    });

    deviceCanvas.mouseleave(function (e) {
        mousePressed = false;
    });

    $('#fileuploader').uploadFile({
        url: "inspector/upload",
        fileName: "test_target",
        multiple: false,
        dragDrop: true,
        returnType: "json",
        onSuccess: function (files, data, xhr, pd) {
            //files: list of files
            //data: response from server
            //xhr : jquer xhr object
            testAppPath = data.app_path;
            $('#appSelectBoxValue').val(testAppPath)
        },
    });
    updateState();
    var sessionId = getCookie("session_id");
    if (sessionId !== "") {
        getScreenshot();
    }

    setInterval(keepAlive, 30000);
})


function selectNode(nodeID) {
    selectedNode = nodeID;
    showElementInfos();
}

function getNodeName(node) {
    var name = node.getAttribute("name");
    if (name.length > 0) {
        name = "(" + name + ")"
    } else {
        name = node.getAttribute("LabelText");
        if (name.length > 0) {
            name = "(" + name + ")"
        }
    }
    return node.nodeName + name
}

function graphRenderChildren(node, expand) {
    if (node.nodeType === 3) {
        return "";
    }
    var html = "";
    if (node.childNodes.length > 1) {
        html = "<li role=\"treeitem\" aria-expanded=\"" + expand + "\">";
        html += "<span onclick='selectNode(\"" + node.getAttribute("id") + "\")'>" + getNodeName(node) + "</span>";
        html += "<ul>";
        var i = 0;
        for (i = 0; i < node.childNodes.length; i++) {
            html += graphRenderChildren(node.childNodes[i], "false");
        }
        if (node.childNodes.length > 0) {
            html += "</ul>";
        }
    } else {
        html = "<li class=\"doc\" onclick='selectNode(\"" + node.getAttribute("id") + "\")'>";
        html += getNodeName(node);
    }
    html += "</li>";
    return html;
}

function updateGraph() {
    var graphViewContainer = $('#screenGraphView');
    var content = "<ul role=\"tree\" aria-labelledby=\"tree_label\">";
    content += graphRenderChildren(xmlDocumentCache.documentElement, "true");
    content += "</ul>";
    graphViewContainer.html(content);
    updateTreeItems();
}