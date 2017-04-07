import Vue from 'vue'
import App from './App.vue'
import Views from './views/_index'
import Components from './components/_index'
import router from './router/index'
import store from './store/index'
import { sync } from 'vuex-router-sync'
import Vuetify from 'vuetify'
import VueRouter from 'vue-router'
import StatsBackend from './statsbackend'

// Open a connection to the Stats Backend
const sb = new StatsBackend()
sb.open()

// Register Components
Object.keys(Components).forEach(key => {
  Vue.component(key, Components[key])
})

sync(store, router)

Vue.use(Vuetify)

new Vue({
  router: router,
  store: store,
  el: '#app',
  render: h => h(App)
})