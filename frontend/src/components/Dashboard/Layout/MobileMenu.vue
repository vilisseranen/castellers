<template>
  <ul class="nav nav-mobile-menu">
    <li class="nav-item" v-if="this.uuid">
      <router-link class="nav-link" :to="{ name: 'MemberEdit', params: { uuid: uuid}}" @click.native="hideSidebar">
        <i class="nc-icon nc-circle-09 hidden-lg-up"></i>
        <span> {{ $t('general.profile_zone') }}</span>
      </router-link>
    </li>
    <li class="nav-item" v-if="!this.uuid">
      <router-link class="nav-link" :to="{ name: 'Login'}" @click.native="hideSidebar">
        <i class="nc-icon nc-tap-01 hidden-lg-up"></i>
        <span> {{ $t('general.login') }}</span>
      </router-link>
    </li>
  </ul>
</template>

<i18n src='assets/translations/general.json'></i18n>

<script>
  import DropDown from 'src/components/UIComponents/Dropdown.vue'
  import {mapGetters} from 'vuex'

  export default {
    inject: {
      autoClose: {
        default: true
      }
    },
    computed: {
      ...mapGetters(['uuid', 'code', 'type'])
    },
    components: {
      DropDown
    },
    methods: {
      hideSidebar () {
        if (this.autoClose && this.$sidebar && this.$sidebar.showSidebar === true) {
          this.$sidebar.displaySidebar(false)
        }
      }
    }
  }
</script>
<style>
</style>
