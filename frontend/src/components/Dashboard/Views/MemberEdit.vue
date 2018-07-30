<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-md-12">
          <edit-profile-form :user="user" :updating="updating">
            <template slot="message">
              <span></span>
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
import {mapGetters} from 'vuex'

export default {
  components: {
    EditProfileForm
  },
  data () {
    return {
      user: {roles: []},
      updating: false
    }
  },
  mounted () {
    if (this.$route.params.uuid !== undefined) {
      var self = this
      axios.get(
        `/api/admins/${this.uuid}/members/${this.$route.params.uuid}`,
        { headers: { 'X-Member-Code': this.code } }
      ).then(function (response) {
        console.log(response.data.roles)
        self.user = response.data
      }).catch(err => console.log(err))
    }
  },
  computed: {
    ...mapGetters(['uuid', 'code', 'type'])
  }
}
</script>
<style>
</style>
