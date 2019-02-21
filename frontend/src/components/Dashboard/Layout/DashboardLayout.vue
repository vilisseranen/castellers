<template>
  <div class="wrapper">
    <side-bar>
      <mobile-menu slot="content"></mobile-menu>
      <sidebar-link to="/members" v-if="type==='admin'">
        <i class="nc-icon nc-bullet-list-67"></i>
        <p>{{ $t('routes.members') }}</p>
      </sidebar-link>
      <sidebar-link to="/practices">
        <i class="nc-icon nc-refresh-02"></i>
        <p>{{ $t('routes.practices') }}</p>
      </sidebar-link>
      <!--
      <sidebar-link to="/events">
        <i class="nc-icon nc-pin-3"></i>
        <p>Events</p>
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

<i18n src='assets/translations/routes.json'></i18n>

<script>
  import TopNavbar from './TopNavbar.vue'
  import ContentFooter from './ContentFooter.vue'
  import DashboardContent from './Content.vue'
  import MobileMenu from './MobileMenu.vue'
  import {mapGetters} from 'vuex'
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
      }
    },
    computed: {
      ...mapGetters(['uuid', 'code', 'type'])
    }
  }

</script>
