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
                    label="First Name"
                    v-model="user.firstName">
          </fg-input>
        </div>
        <div class="col-md-4">
          <fg-input type="text"
                    label="Last Name"
                    v-model="user.lastName">
          </fg-input>
        </div>
        <div class="col-md-4">
          <fg-input type="email"
                    label="Email"
                    placeholder="Email"
                    v-model="user.email">
          </fg-input>
        </div>
      </div>
      <div class="row">
        <div class="col-md-6">
          <fg-input type="text"
                    label="Roles"
                    placeholder="Baix"
                    v-model="user.roles">
          </fg-input>
        </div>
        <div class="col-md-6">
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
      <div class="clearfix"></div>
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
    user: Object
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
