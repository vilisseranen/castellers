<template>
  <div class="wrapper">
    <side-bar>
      <mobile-menu slot="content"></mobile-menu>
      <sidebar-link to="/news">
        <i class="nc-icon nc-attach-87"></i>
        <p>News</p>
      </sidebar-link>
      <sidebar-link to="/practices">
        <i class="nc-icon nc-refresh-02"></i>
        <p>Practices</p>
      </sidebar-link>
      <sidebar-link to="/events">
        <i class="nc-icon nc-pin-3"></i>
        <p>Events</p>
      </sidebar-link>
      <!-- TODO: add a gallery
      <sidebar-link to="/admin/typography">
        <i class="nc-icon nc-satisfied"></i>
        <p>Gallery</p>
      </sidebar-link>
      -->
    </side-bar>
    <div class="main-panel">
      <top-navbar></top-navbar>

      <dashboard-content @click="toggleSidebar">

      </dashboard-content>

      <content-footer></content-footer>
    </div>
  </div>
</template>
<style lang="scss">

</style>
<script>
  import TopNavbar from './TopNavbar.vue'
  import ContentFooter from './ContentFooter.vue'
  import DashboardContent from './Content.vue'
  import MobileMenu from './MobileMenu.vue'
  export default {
    components: {
      TopNavbar,
      ContentFooter,
      DashboardContent,
      MobileMenu
    },
    methods: {
      toggleSidebar () {
        if (this.$sidebar.showSidebar) {
          this.$sidebar.displaySidebar(false)
        }
      },
      sayHello () {
      }

    },
    mounted () {
      // this.sayHello()
    }
  }

</script>
