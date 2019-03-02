<template>
  <div class="content">
    <div class="container-fluid">
      <card>
        <template slot="header">
          <h4 class="card-title">{{ $t('login.status') }}</h4>
        </template>
        <div class="row">
          <div class="col-md">
            <div class="alert alert-warning" v-if="!uuid || !code">
              <span><b>{{ $t('login.not_logged_in') }}</b> - {{ $t('login.not_logged_in_text') }}</span>
            </div>
            <div class="alert alert-success" v-if="uuid && code">
              <span><b>{{ $t('login.logged_in') }}</b> - {{ $t('login.logged_in_text') }}</span>
            </div>
          </div>
        </div>
      </card>
      <login-card v-if="!uuid || !code">
      </login-card>
      <card v-if="uuid && code">
        <template slot="header">
          <h4 class="card-title">{{ $t('login.now') }}</h4>
        </template>
        <div class="row">
          <div class="col-md">
            <h5>{{ $t('login.logged_in_instructions') }}</h5>
            <ul>
              <li>{{ $t('login.bookmark_site') }}</li>
              <li>{{ $t('login.update_profile') }}</li>
              <li>{{ $t('login.register_events') }}</li>
            </ul>
          </div>
        </div>
      </card>
      <card v-if="uuid && code">
        <h4 slot="header" class="card-title">{{ $t('login.autoconnect') }}</h4>
        <PrettyCheck class="p-default p-curve" v-model="autoconnect">{{ $t('login.autoconnect_' + autoconnectLabel) }}</PrettyCheck>
        <div slot="footer">
          <p>{{ $t('login.cookie_warning') }}</p>
        </div>
      </card>
    </div>
  </div>
</template>

<i18n src='assets/translations/login.json'></i18n>

<script>
  import {mapGetters} from 'vuex'
  import Card from 'src/components/UIComponents/Cards/Card.vue'
  import LoginCard from 'src/components/UIComponents/Cards/LoginCard.vue'
  import PrettyCheck from 'pretty-checkbox-vue/check'
  import {cookieMixin} from 'src/components/mixins/cookies.js'

  import 'pretty-checkbox/dist/pretty-checkbox.min.css'

  export default {
    mixins: [cookieMixin],
    components: {
      Card,
      LoginCard,
      PrettyCheck
    },
    computed: {
      ...mapGetters(['uuid', 'code', 'type']),
      autoconnectLabel: function () {
        return this.autoconnect ? 'yes' : 'no'
      }
    },
    data () {
      return {
        autoconnect: false
      }
    },
    watch: {
      autoconnect: function (val) {
        if (val) {
          this.setCookie('member', this.uuid, 365)
          this.setCookie('code', this.code, 365)
        } else {
          this.eraseCookie('member')
          this.eraseCookie('code')
        }
      }
    },
    mounted () {
      var member = this.getCookie('member')
      var code = this.getCookie('code')
      if (member && code) {
        this.autoconnect = true
      }
    }
  }

</script>
<style lang="scss">

</style>
