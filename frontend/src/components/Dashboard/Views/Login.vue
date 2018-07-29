<template>
  <div class="content">
    <div class="container-fluid">
      <card>
        <template slot="header">
          <h4 class="card-title">Login</h4>
        </template>
        <div class="row">
          <div class="col-md">
            <div class="alert alert-warning" v-if="!uuid || !code">
              <span><b> You are not authenticated</b> If you want to register for events or check your profile, please enter your member code.</span>
            </div>
            <div class="alert alert-success" v-if="uuid && code">
              <span><b> You are already authenticated</b> You will be able to register for events.</span>
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
  import {mapGetters} from 'vuex'
  import Card from 'src/components/UIComponents/Cards/Card.vue'
  import LoginCard from 'src/components/UIComponents/Cards/LoginCard.vue'

  export default {
    components: {
      Card,
      LoginCard
    },
    computed: {
      ...mapGetters(['uuid', 'code', 'type'])
    }
  }

</script>
<style lang="scss">

</style>
