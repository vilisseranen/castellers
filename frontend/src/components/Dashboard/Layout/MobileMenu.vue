<template>
  <ul class="nav nav-mobile-menu">
    <li class="nav-item">
      <router-link class="nav-link" :to="{ name: 'MemberEdit', params: { uuid: uuid}}">
        <i class="nc-icon nc-circle-09 hidden-lg-up"></i>
        <span> {{ $t('general.profile_zone') }}</span>
      </router-link>
    </li>
  </ul>
</template>

<i18n src='assets/translations/general.json'></i18n>

<script>
  import DropDown from 'src/components/UIComponents/Dropdown.vue'
  import {mapGetters} from 'vuex'

  export default {
    computed: {
      ...mapGetters(['uuid', 'code', 'type'])
    },
    components: {
      DropDown
    }
  }
</script>
<style>
</style>
