<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-md-12">
          <edit-profile-form :user="user" :updating="updating" v-on:updateUser="loadUser" v-on:deleteUser="removeUser">
            <template slot="message">
              <span></span>
            </template>
          </edit-profile-form>
        </div>
      </div>
    </div>
  </div>
</template>

<i18n src='assets/translations/members.json'></i18n>

<script>
import EditProfileForm from './UserProfile/EditProfileForm.vue'
import axios from 'axios'
import {mapGetters} from 'vuex'
import {memberMixin} from 'src/components/mixins/members.js'

export default {
  mixins: [memberMixin],
  components: {
    EditProfileForm
  },
  data () {
    return {
      user: {roles: [], type: 'member'},
      updating: false
    }
  },
  mounted () {
    this.loadUser(this.$route.params.uuid)
  },
  computed: {
    ...mapGetters(['uuid', 'code', 'type'])
  },
  methods: {
    loadUser (uuid) {
      if (uuid) {
        var self = this
        axios.get(
          `/api/admins/${this.uuid}/members/${uuid}`,
          { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.user = response.data
          self.$router.push({path: `/memberEdit/${self.user.uuid}`})
        }).catch(err => console.log(err))
      }
    },
    removeUser (member) {
      var self = this
      this.deleteUser(member)
        .then(function() { self.$router.push({path: `/members`})})
        .catch(function(error) { console.log(error) })
    }
  }
}
</script>
<style>
</style>
