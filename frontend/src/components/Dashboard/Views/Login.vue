<template>
  <div class="content">
    <div class="container-fluid">
      <card>
        <template slot="header">
          <h4 class="card-title">Login</h4>
        </template>
        <div class="row">
          <div class="col-md">
            <div class="alert alert-warning">
              <span><b> You are not authenticated</b> If you want to register for events or check your profile, please enter your member code.</span>
            </div>
          </div>
        </div>
      </card>
      <login-card>
      </login-card>
    </div>
  </div>
</template>
<script>
  import Card from 'src/components/UIComponents/Cards/Card.vue'
  import LoginCard from 'src/components/UIComponents/Cards/LoginCard.vue'

  export default {
    components: {
      Card,
      LoginCard
    },
    data () {
      return {
        type: ['', 'info', 'success', 'warning', 'danger'],
        notifications: {
          topCenter: false
        }
      }
    },
    methods: {
      notifyVue (verticalAlign, horizontalAlign) {
        const notification = {
          template: `<span>Welcome to <b>Light Bootstrap Dashboard</b> - a beautiful freebie for every web developer.</span>`
        }

        const color = Math.floor((Math.random() * 4) + 1)
        this.$notifications.notify(
          {
            component: notification,
            icon: 'nc-icon nc-app',
            horizontalAlign: horizontalAlign,
            verticalAlign: verticalAlign,
            type: this.type[color]
          })
      }
    }
  }

</script>
<style lang="scss">

</style>
