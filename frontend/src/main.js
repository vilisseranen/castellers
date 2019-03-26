import Vue from 'vue'
import VueRouter from 'vue-router'
import Vuex from 'vuex'
import App from './App.vue'
import axios from 'axios'
import VueI18n from 'vue-i18n'
import VuejsDialog from 'vuejs-dialog'
import ToggleButton from 'vue-js-toggle-button'

import 'vuejs-dialog/dist/vuejs-dialog.min.css'

import {cookieMixin} from 'src/components/mixins/cookies.js'

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
Vue.use(ToggleButton)
Vue.use(cookieMixin)

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
    action: {
      type: '',
      objectUUID: '',
      payload: ''
    },
    locale: ''
  },
  mutations: {
    authenticate (state, payload) {
      state.auth.uuid = payload.uuid
      state.auth.code = payload.code
      state.auth.type = payload.type
      state.locale = payload.language
    },
    setAction (state, payload) {
      state.action.type = payload.type
      state.action.objectUUID = payload.objectUUID
      state.action.payload = payload.payload
    }
  },
  getters: {
    uuid: (state) => state.auth.uuid,
    code: (state) => state.auth.code,
    type: (state) => state.auth.type,
    language: (state) => state.locale,
    action: (state) => state.action
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
  mixins: [cookieMixin],
  methods: {
    setLocale: function (locale) {
      this.$i18n.locale = locale
    },
    globalRedirect () {
      if ('next' in this.$route.query) {
        this.$router.push(this.$route.query.next)
      }
    },
    checkAction () {
      if ('action' in this.$route.query) {
        var action = {}
        action.type = this.$route.query.action
        action.objectUUID = this.$route.query.objectUUID
        action.payload = this.$route.query.payload
        this.$store.commit('setAction', action)
      }
    },
    checkCredentials () {
      var self = this
      var cookieMember = this.getCookie('member')
      var cookieCode = this.getCookie('code')
      if (cookieMember && cookieCode) {
        axios.get(
          '/api/members/' + cookieMember,
          { headers: { 'X-Member-Code': cookieCode } }
        ).then(function (response) {
          self.$store.commit('authenticate', {
            uuid: response.data.uuid,
            code: cookieCode,
            type: response.data.type,
            language: response.data.language
          })
          self.setLocale(response.data.language)
          self.globalRedirect()
        }).catch(err => {
          console.log(err)
        })
      } else if ('c' in this.$route.query && 'm' in this.$route.query) {
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
    this.checkAction()      // check for an action to perform
    this.checkCredentials() // log in and go to next page
  }
})
