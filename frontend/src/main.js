import Vue from 'vue'
import VueRouter from 'vue-router'
import App from './App.vue'

// LightBootstrap plugin
import LightBootstrap from './light-bootstrap-main'

// router setup
import routes from './routes/routes'
// plugin setup
Vue.use(VueRouter)
Vue.use(LightBootstrap)

// configure router
const router = new VueRouter({
  routes, // short for routes: routes
  linkActiveClass: 'nav-item active'
})

/* eslint-disable no-new */
new Vue({
  el: '#app',
  render: h => h(App),
  router,
  methods: {
    globalRedirect () {
      if ('next' in this.$route.query) {
        this.$router.push(this.$route.query.next)
        console.log("Redirecting to " + this.$route.query.next)
      }
    },
    checkCode() {
      if ('code' in this.$route.query) {
        console.log("You provided the code: " + this.$route.query.code)
      }
    }
  },
  mounted() {
    this.globalRedirect()
    this.checkCode()
  }
})