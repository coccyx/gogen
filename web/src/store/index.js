import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

export default new Vuex.Store({
  state: {
    title: null,
    description: null,
    keywords: null
  },

  actions: {},

  mutations: {
    'gogen/TITLE'(state, payload) {
      state.title = payload
      document.title = payload
    }
  }
})