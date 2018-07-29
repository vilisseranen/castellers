<template>
  <card>
    <h4 slot="header" class="card-title">Enter member ID</h4>
    <form>
      <div class="row">
        <div class="col-md-12">
          <fg-input type="text"
                    placeholder="335b9fba95a1578baa5a2b9560e3566f174ed3a7"
                    v-model="member.uuid">
          </fg-input>
        </div>
      </div>
      <h4 slot="header" class="card-title">Enter member code</h4>
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
          Login
          </button>
        </div>
      </div>
      <h4 class="text-primary">
        You can also simply click on the link you received by email ;-)
      </h4>
    </form>
  </card>
</template>
<script>
  import Card from 'src/components/UIComponents/Cards/Card.vue'
  import {mapMutations} from 'vuex'
  import axios from 'axios'

  export default {
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
        console.log('uuid: ' + this.member.uuid)
        console.log('code: ' + this.member.code)
        axios.get(
          '/api/members/' + this.member.uuid,
          { headers: { 'X-Member-Code': this.member.code } }
          ).then(function (response) {
            console.log(self.member)
            self.member.type = response.data.type
            self.authenticate(self.member)
          }).catch(err => {
            console.log(err)
            self.notifyNOK()
          })
      },
      notifyNOK () {
        const notification = {
          template: `<span>There was an error during login.</span>`
        }
        this.$notifications.notify({
          component: notification,
          icon: 'nc-icon nc-simple-remove',
          type: 'danger',
          showClose: false,
          timeout: null
        })
      }
    }
  }

</script>
<style>

</style>
