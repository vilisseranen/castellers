<template>
  <card>
    <h4 slot="header" class="card-title">Enter member code</h4>
    <form>
      <div class="row">
        <div class="col-md-10">
          <fg-input type="text"
                    placeholder="335b9fba95a1578baa5a2b9560e3566f174ed3a7"
                    v-model="member.uuid">
          </fg-input>
        </div>
        <div class="col-md-2">
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

  export default {
    components: {
      Card
    },
    data () {
      return {
        member: {
          uuid: ''
        }
      }
    },
    methods: {
      login () {
        alert('Your uuid: ' + JSON.stringify(this.member.uuid))
      }
    }
  }

</script>
<style>

</style>
