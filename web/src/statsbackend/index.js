import WebSocketClient from './websocketclient'
import store from '../store'

class StatsBackend {
    open() {
        let wsc = new WebSocketClient();
        wsc.open('ws://localhost:9999/statsws');
        wsc.onerror = function (e) {
            store.commit('backendError', {
                backendError: true,
                backendErrorMsg: "Error communicating with the backend at ws://localhost:9999/statsws"
            })
        }
        wsc.onmessage = function (data, flags, number) {
            let parsed = JSON.parse(data.data)
            if ('EventsWritten' in parsed) {
                store.commit('outputstats', parsed)
            } else if ('GeneratorQueueDepth' in parsed) {
                store.commit('queuedepthstats', parsed)
            }
        }
        this.wsc = wsc;
    }
}

export default StatsBackend