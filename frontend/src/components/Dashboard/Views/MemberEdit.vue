<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-md-12">
          <edit-profile-form :user="user" :updating="updating" v-on:updateUser="loadUser">
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
    this.loadUser()
  },
  computed: {
    ...mapGetters(['uuid', 'code', 'type'])
  },
  methods: {
    loadUser () {
      if (this.$route.params.uuid !== undefined) {
      var self = this
      axios.get(
        `/api/admins/${this.uuid}/members/${this.$route.params.uuid}`,
        { headers: { 'X-Member-Code': this.code } }
      ).then(function (response) {
        self.user = response.data
      }).catch(err => console.log(err))
    }
    }
  }
}
</script>
<style>
</style>
