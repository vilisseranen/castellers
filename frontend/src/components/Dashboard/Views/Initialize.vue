<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-md-12">
          <edit-profile-form>
            <template slot="update-button">
              <button type="submit" class="btn btn-info btn-fill float-right" @click.prevent="initializeApp">
                Initialize app
              </button>
            </template>
          </edit-profile-form>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import EditProfileForm from './UserProfile/EditProfileForm.vue'
import axios from 'axios'

export default {
  components: {
    EditProfileForm
  },
  methods: {
    initializeApp () {
      axios.post('/initialize', {'name': 'ian', 'extra': 'Cap de colla'}).then(function (response) {
        alert(response)
      })
    }
  }
}
</script>
<style>
</style>
