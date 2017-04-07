function WebSocketClient() {
    this.number = 0;    // Message number
    this.autoReconnectInterval = 5 * 1000;    // ms
}

const WSDEBUG = 0

WebSocketClient.prototype.open = function (url) {
    this.url = url;
    this.instance = new WebSocket(this.url);
    this.instance.onopen
    this.instance.onopen = (e) => {
        this.onopen(e);
    };
    this.instance.onmessage = (data, flags) => {
        this.number++;
        this.onmessage(data, flags, this.number);
    };
    this.instance.onclose = (e) => {
        switch (e) {
            case 1000:  // CLOSE_NORMAL
                if (WSDEBUG) {
                    console.log("WebSocket: closed");
                }
                break;
            default:    // Abnormal closure
                this.reconnect(e);
                break;
        }
        this.onclose(e);
    };
    this.instance.onerror = (e) => {
        switch (e.code) {
            case 'ECONNREFUSED':
                this.reconnect(e);
                break;
            default:
                this.onerror(e);
                break;
        }
    };
}
WebSocketClient.prototype.send = function (data, option) {
    try {
        this.instance.send(data, option);
    } catch (e) {
        this.instance.emit('error', e);
    }
}
WebSocketClient.prototype.reconnect = function (e) {
    if (WSDEBUG) {
        console.log(`WebSocketClient: retry in ${this.autoReconnectInterval}ms`, e);
    }
    var that = this;
    setTimeout(function () {
        if (WSDEBUG) {
            console.log("WebSocketClient: reconnecting...");
        }
        that.open(that.url);
    }, this.autoReconnectInterval);
}
WebSocketClient.prototype.onopen = function (e) { if (WSDEBUG) { console.log("WebSocketClient: open", arguments); } }
WebSocketClient.prototype.onmessage = function (data, flags, number) { if (WSDEBUG) { console.log("WebSocketClient: message", arguments); } }
WebSocketClient.prototype.onerror = function (e) { if (WSDEBUG) { console.log("WebSocketClient: error", arguments); } }
WebSocketClient.prototype.onclose = function (e) { if (WSDEBUG) { console.log("WebSocketClient: closed", arguments); } }

export default WebSocketClient