import Vue from 'vue'
import VueRouter from 'vue-router'
import Vuex from 'vuex'
import App from './App.vue'
import axios from 'axios'
import VueI18n from 'vue-i18n'
import VuejsDialog from 'vuejs-dialog'

import 'vuejs-dialog/dist/vuejs-dialog.min.css'

// LightBootstrap plugin
import LightBootstrap from './light-bootstrap-main'

// router setup
import routes from './routes/routes'

// plugin setup
Vue.use(VueRouter)
Vue.use(LightBootstrap)
Vue.use(Vuex)
Vue.use(VueI18n)
Vue.use(VuejsDialog)

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
    },
    locale: ''
  },
  mutations: {
    authenticate (state, payload) {
      state.auth.uuid = payload.uuid
      state.auth.code = payload.code
      state.auth.type = payload.type
      state.locale = payload.language
    }
  },
  getters: {
    uuid: (state) => state.auth.uuid,
    code: (state) => state.auth.code,
    type: (state) => state.auth.type,
    language: (state) => state.locale
  }
})

// Configure i18n
const i18n = new VueI18n({
  locale: 'fr', // set locale
  fallbackLocale: 'fr',
  messages: {}
})

/* eslint-disable no-new */
new Vue({
  el: '#app',
  render: h => h(App),
  router,
  store,
  i18n,
  methods: {
    setLocale: function (locale) {
      this.$i18n.locale = locale
    },
    globalRedirect () {
      if ('next' in this.$route.query) {
        this.$router.push(this.$route.query.next)
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
            uuid: response.data.uuid,
            code: self.$route.query.c,
            type: response.data.type,
            language: response.data.language
          })
          self.setLocale(response.data.language)
          self.globalRedirect()
        }).catch(err => {
          console.log(err)
        })
      } else {
        self.globalRedirect()
      }
    }
  },
  mounted () {
    this.checkCredentials() // this must be the first thing to do
  }
})
