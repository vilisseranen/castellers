<template>
  <card>
    <h4 slot="header" class="card-title">Edit Profile</h4>
    <form>
      <div class="row">
        <div class="col-md-12">
          <fg-input type="text"
                    label="ID"
                    :disabled="true"
                    v-model="user.uuid">
          </fg-input>
        </div>
      </div>
      <div class="row">
        <div class="col-md-4">
          <fg-input type="text"
                    label="Name"
                    placeholder="Name"
                    v-model="user.name">
          </fg-input>
        </div>
        <div class="col-md-4">
          <fg-input type="email"
                    label="Email"
                    placeholder="Email"
                    v-model="user.email">
          </fg-input>
        </div>
        <div class="col-md-4">
          <fg-input type="text"
                    label="Extra"
                    placeholder="Extra"
                    v-model="user.extra">
          </fg-input>
        </div>
      </div>
      <div class="text-center">
        <slot name="update-button">
          <button slot="update_button" type="submit" class="btn btn-info btn-fill float-right" @click.prevent="updateProfile">
            Update Profile
          </button>
        </slot>
      </div>
      <div class="clearfix">
        <div class="spinner" v-if="updating == true">
          <div class="double-bounce1"></div>
          <div class="double-bounce2"></div>
        </div>
      </div>
    </form>
  </card>
</template>
<script>
import Card from 'src/components/UIComponents/Cards/Card.vue'

export default {
  components: {
    Card
  },
  name: 'edit-profile-form',
  props: {
    user: Object,
    updating: Boolean
  },
  methods: {
    updateProfile () {
      alert('Your data: ' + JSON.stringify(this.user))
    }
  }
}
</script>
<style>
</style>
