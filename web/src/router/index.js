import Vue from 'vue'
import VueRouter from 'vue-router'

const routes = [
    { path: '/', component: require('../views/Home.vue') },
    { path: '/editor', component: require('../views/Editor.vue') },
    { path: '/stats', component: require('../views/Stats.vue') }
]

Vue.use(VueRouter)

const router = new VueRouter({
    routes: routes
})

export default router