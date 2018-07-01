<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-md-12">
          <edit-profile-form :user="user" :updating="updating" v-if="initialized == false">
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
    var updating = false
    var user = {}
    axios.get('http://127.0.0.1:8080/initialize').then(function (response) {
      if (response.status === 204) {
        self.initialized = false
      } else {
        alert('app already initialized')
      }
    })
    return {
      user,
      initialized,
      updating
    }
  },
  methods: {
    initializeApp () {
      var self = this
      self.updating = true
      axios.post('http://127.0.0.1:8080/initialize', self.user).then(function (response) {
        self.updating = false
        self.user = response.data
        if (response.status === 201) {
          self.notifyOK()
        } else {
          self.notifyNOK()
        }
      }).catch(err => console.log(err))
    },
    notifyOK () {
      const notification = {
        template: `<span>The application is now initialized ! You will receive an email with your infos.</span>`
      }
      this.$notifications.notify({
        component: notification,
        icon: 'nc-icon nc-check-2',
        type: 'success',
        timeout: null
      })
    },
    notifyNOK () {
      const notification = {
        template: `<span>There was an error during the application initialization.</span>`
      }
      this.$notifications.notify({
        component: notification,
        icon: 'nc-icon nc-simple-remove',
        type: 'danger',
        timeout: null
      })
    }
  }
}
</script>
<style>
</style>
