<template>
  <nav class="navbar navbar-expand-lg">
    <div class="container-fluid">
      <a class="navbar-brand" href="#">{{ $t(routeName + '.category') }}</a>
      <button type="button"
              class="navbar-toggler navbar-toggler-right"
              :class="{toggled: $sidebar.showSidebar}"
              aria-controls="navigation-index"
              aria-expanded="false"
              aria-label="Toggle navigation"
              @click="toggleSidebar">
        <span class="navbar-toggler-bar burger-lines"></span>
        <span class="navbar-toggler-bar burger-lines"></span>
        <span class="navbar-toggler-bar burger-lines"></span>
      </button>
      <div class="collapse navbar-collapse justify-content-end">
        <ul class="navbar-nav ml-auto" v-if="this.$store.state.auth.uuid">
          <li class="nav-item" v-if="this.$store.state.auth.type === 'admin'">
            <a class="nav-link" href="#">
              Admin
            </a>
          </li>
          <li class="nav-item">
            <a href="#" class="nav-link">
              Profile
            </a>
          </li>
        </ul>
      </div>
    </div>
  </nav>
</template>

<i18n src='assets/translations/members.json'></i18n>

<script>
  export default {
    computed: {
      routeName () {
        const {path} = this.$route
        return path.substring(1) // This removes the leading '/'
      }
    },
    data () {
      return {
        activeNotifications: false
      }
    },
    methods: {
      capitalizeFirstLetter (string) {
        return string.charAt(0).toUpperCase() + string.slice(1)
      },
      toggleNotificationDropDown () {
        this.activeNotifications = !this.activeNotifications
      },
      closeDropDown () {
        this.activeNotifications = false
      },
      toggleSidebar () {
        this.$sidebar.displaySidebar(!this.$sidebar.showSidebar)
      },
      hideSidebar () {
        this.$sidebar.displaySidebar(false)
      }
    }
  }

</script>
<style>

</style>
