import Vue from 'vue'
import VueRouter from 'vue-router'
import Vuex from 'vuex'
import App from './App.vue'

// LightBootstrap plugin
import LightBootstrap from './light-bootstrap-main'

// router setup
import routes from './routes/routes'
// plugin setup
Vue.use(VueRouter)
Vue.use(LightBootstrap)
Vue.use(Vuex)

// configure router
const router = new VueRouter({
  routes, // short for routes: routes
  linkActiveClass: 'nav-item active'
})

// Configure vuex store
const store = new Vuex.Store({
  state: {
    auth: {
      uuid: '',
      code: '',
      type: ''
    }
  },
  mutations: {
    authenticate (state, payload) {
      state.auth.uuid = payload.uuid
      state.auth.code = payload.code
      state.auth.type = payload.type
    }
  }
})

/* eslint-disable no-new */
new Vue({
  el: '#app',
  render: h => h(App),
  router,
  store,
  methods: {
    globalRedirect () {
      if ('next' in this.$route.query) {
        this.$router.push(this.$route.query.next)
        console.log('Redirecting to ' + this.$route.query.next)
      }
    },
    checkCredentials () {
      if ('c' in this.$route.query && 'm' in this.$route.query) {
        this.$store.commit('authenticate', {
          uuid: this.$route.query.m,
          code: this.$route.query.c,
          type: 'admin'
        })
        console.log('You are authenticated as : ' + JSON.stringify(this.$store.state.auth))
      }
    }
  },
  mounted () {
    this.checkCredentials() // this must be the first thing to do
    this.globalRedirect()   // you redirect after being authenticated
  }
})
