import Vue from 'vue'
import VueRouter from 'vue-router'
import Vuex from 'vuex'
import App from './App.vue'
import axios from 'axios'

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
  },
  getters: {
    uuid: (state) => state.auth.uuid,
    code: (state) => state.auth.code,
    type: (state) => state.auth.type
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
      var self = this
      if ('c' in this.$route.query && 'm' in this.$route.query) {
        axios.get(
          '/api/members/' + this.$route.query.m,
          { headers: { 'X-Member-Code': this.$route.query.c } }
        ).then(function (response) {
          self.$store.commit('authenticate', {
            uuid: self.$route.query.m,
            code: self.$route.query.c,
            type: response.data.type
          })
          console.log('You are authenticated as : ' + JSON.stringify(self.$store.state.auth))
        }).catch(err => {
          console.log(err)
        })
      }
    }
  },
  mounted () {
    this.checkCredentials() // this must be the first thing to do
    this.globalRedirect()   // you redirect after being authenticated
  }
})
