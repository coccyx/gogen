import Vue from 'vue'
import App from './App.vue'
import Views from './views/_index'
import router from './router/index'
import store from './store/index'
import { sync } from 'vuex-router-sync'
import Vuetify from 'vuetify'
import VueRouter from 'vue-router'

sync(store, router)

Vue.use(Vuetify)

new Vue({
  router: router,
  store: store,
  el: '#app',
  render: h => h(App)
})