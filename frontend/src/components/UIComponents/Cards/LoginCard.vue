<template>
  <card>
    <h4 slot="header">{{ $t('login.how_to_log_in') }}</h4>
    <p>{{ $t('login.instructions') }}</p>
    <h4 class="card-title">{{ $t('login.id') }}</h4>
    <form>
      <div class="row">
        <div class="col-md-12">
          <fg-input type="text"
                    placeholder="335b9fba95a1578baa5a2b9560e3566f174ed3a7"
                    v-model="member.uuid">
          </fg-input>
        </div>
      </div>
      <h4 class="card-title">{{ $t('login.code') }}</h4>
      <div class="row">
        <div class="col-md-12">
          <fg-input type="text"
                    placeholder="335b9fba95a1"
                    v-model="member.code">
          </fg-input>
        </div>
      </div>
      <div class="row">
        <div class="col-md-12">
          <button type="submit" class="btn btn-info btn-fill float-right" @click.prevent="login">
          {{ $t('login.button') }}
          </button>
        </div>
      </div>
    </form>
  </card>
</template>

<i18n src='assets/translations/login.json'></i18n>

<script>
  import Card from 'src/components/UIComponents/Cards/Card.vue'
  import {mapMutations} from 'vuex'
  import axios from 'axios'
  import {notificationMixin} from 'src/components/mixins/notifications.js'

  export default {
    mixins: [notificationMixin],
    components: {
      Card
    },
    data () {
      return {
        member: {}
      }
    },
    methods: {
      ...mapMutations({
        authenticate: 'authenticate'
      }),
      login () {
        var self = this
        axios.get(
          '/api/members/' + this.member.uuid,
          { headers: { 'X-Member-Code': this.member.code } }
          ).then(function (response) {
            self.member.type = response.data.type
            self.authenticate(self.member)
            self.$root.setLocale(response.data.language)
          }).catch(err => {
            console.log(err)
            self.notifyNOK(self.$t('login.notify_error'))
          })
      }
    }
  }

</script>
<style>

</style>
