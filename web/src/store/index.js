import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)


export default new Vuex.Store({
  state: {
    title: null,
    description: null,
    keywords: null,
    totalBytes: 0,
    totalEvents: 0,
    totalSources: 0,
    eventsPerSec: 0,
    bytesPerSec: 0,
    seenSources: {},
    currentGeneratorQueue: 0,
    currentOutputQueue: 0,
    outputStats: [],
    queueDepthStats: [],
    totalOutputStatsMessages: 0,
    backendError: false,
    backendErrorMsg: "",
    lastOutputStatsTS: 0,
    lastQueueDepthStatsTS: 0
  },

  actions: {},

  mutations: {
    title(state, payload) {
      state.title = payload
      document.title = payload
    },
    outputstats(state, stats) {
      state.totalOutputStatsMessages += 1
      state.totalBytes += stats['BytesWritten']
      state.totalEvents += stats['EventsWritten']
      if (!(stats['Source'] in state.seenSources)) {
        state.totalSources += 1
        state.seenSources[stats['Source']] = 1
      }
      if (state.lastOutputStatsTS == 0) {
        state.eventsPerSec = stats['EventsWritten']
        state.bytesPerSecd = stats['BytesWritten']
      } else {
        state.eventsPerSec = stats['EventsWritten'] / (stats['Timestamp'] - state.lastOutputStatsTS)
        state.bytesPerSec = stats['BytesWritten'] / (stats['Timestamp'] - state.lastOutputStatsTS)
      }
      state.lastOutputStatsTS = stats['Timestamp']
      state.outputStats.push(stats)
      state.backendError = false
    },
    queuedepthstats(state, stats) {
      state.currentGeneratorQueue = stats['GeneratorQueueDepth']
      state.currentOutputQueue = stats['OutputQueueDepth']
      state.lastQueueDepthStatsTS = stats['Timestamp']
      state.queueDepthStats.push(stats)
      state.backendError = false
    },
    backendError(state, error) {
      state.backendError = error.backendError
      state.backendErrorMsg = error.backendErrorMsg
    }
  }
})