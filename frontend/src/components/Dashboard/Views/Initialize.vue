<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-md-12">
          <edit-profile-form :user="user" v-if="initialized == false">
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
  data () {
    var self = this
    var initialized = true
    var user = {
    }
    axios.get('http://127.0.0.1:8080/initialize').then(function (response) {
      if (response.status === 204) {
        self.initialized = false
      } else {
        alert('app already initialized')
      }
    })
    return {
      user,
      initialized
    }
  },
  methods: {
    initializeApp () {
      axios.post('http://127.0.0.1:8080/initialize', {'name': 'ian', 'extra': 'Cap de colla'}).then(function (response) {
        alert(response)
      })
    }
  }
}
</script>
<style>
</style>
