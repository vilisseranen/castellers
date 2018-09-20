<template>
  <footer class="footer">
    <div class="container-fluid">
      <div class="lang text-center">
        <a href="#" v-bind:class="[ this.$i18n.locale == 'fr' ? 'active' : 'inactive' ]" v-on:click="changeLanguage('fr')">Français</a>
        <a href="#" v-bind:class="[ this.$i18n.locale == 'en' ? 'active' : 'inactive' ]" v-on:click="changeLanguage('en')">English</a>
        <a href="#" v-bind:class="[ this.$i18n.locale == 'cat' ? 'active' : 'inactive' ]" v-on:click="changeLanguage('cat')">Català</a>
      </div>
      <div class="copyright text-center">
        {{ $t('general.for_castellers') }}
        <a href="https://github.com/vilisseranen" target="_blank">Clément Contini</a>.
      </div>
    </div>
  </footer>
</template>

<i18n src='assets/translations/general.json'></i18n>

<script>
  export default {
    methods: {
      changeLanguage: function (lang) {
        this.$root.setLocale(lang)
      }
    }
  }

</script>
<style>

</style>
