<template>
  <card>
    <h4 slot="header" class="card-title">{{actionLabel}}</h4>
    <form>
      <div class="row">
        <div class="col-md-8">
          <fg-input type="text"
                    label="ID"
                    :disabled="true"
                    v-model="current_user.uuid">
          </fg-input>
        </div>
        <div class="col-md-4">
        <fg-input label="type" type="radio">
          <form slot="input" id="test">
              <PrettyRadio class="p-default p-curve" name="type" color="primary-o" value="member" v-model="current_user.type">Member</PrettyRadio>
              <PrettyRadio class="p-default p-curve" name="type" color="success-o" value="admin" v-model="current_user.type">Admin</PrettyRadio>
          </form>
        </fg-input>
        </div>
      </div>
      <div class="row">
        <div class="col-md-4">
          <fg-input type="text"
                    label="First Name"
                    placeholder="First Name"
                    v-model="current_user.firstName">
          </fg-input>
        </div>
        <div class="col-md-4">
          <fg-input type="text"
                    label="Last Name"
                    placeholder="Last Name"
                    v-model="current_user.lastName">
          </fg-input>
        </div>
        <div class="col-md-4">
          <fg-input type="email"
                    label="Email"
                    placeholder="Email"
                    v-model="current_user.email">
          </fg-input>
        </div>
      </div>
      <div class="row">
        <div class="col-md-8">
          <fg-input type="text"
                    label="Roles">
            <template slot="input">
              <multiselect
                v-model="current_user.roles"
                :options="available_roles"
                :multiple="true"
                :placeholder="''"
                :closeOnSelect="false">
              </multiselect>
            </template>
          </fg-input>
        </div>
        <div class="col-md-4">
          <fg-input type="text"
                    label="Extra"
                    placeholder="Extra"
                    v-model="current_user.extra">
          </fg-input>
        </div>
      </div>
      <div slot="message" class="row">
        <div class="col-md-12">
          <div class="alert alert-success" v-if="current_user.activated === 1">
            <span><b> Success - </b> This has logged in.</span>
          </div>
           <div class="alert alert-warning" v-if="current_user.activated === 0">
            <span><b> Warning - </b> This user has not logged in yet.</span>
          </div>
        </div>
      </div>
      <div class="text-center">
        <slot name="update-button">
          <button slot="update_button" type="submit" class="btn btn-info btn-fill float-right" @click.prevent="memberEdit">
            {{actionLabel}}
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
import axios from 'axios'
import {mapGetters} from 'vuex'
import Multiselect from 'vue-multiselect'
import 'vue-multiselect/dist/vue-multiselect.min.css'
import PrettyRadio from 'pretty-checkbox-vue/radio';
import 'pretty-checkbox/dist/pretty-checkbox.min.css'

export default {
  components: {
    Card,
    Multiselect,
    PrettyRadio
  },
  name: 'edit-profile-form',
  props: {
    user: Object
  },
  computed: {
    ...mapGetters(['uuid', 'code', 'type']),
    actionLabel: function () {
      return this.current_user.uuid ? "Update user" : "Create user"
    },
    current_user: {
      get: function () {
        return this.user
      },
      set: function (response) {
        return response
      }
    }
  },
  data () {
    return {
      updating: false,
      available_roles: []
    }
  },
  mounted () {
    var self = this
    this.selected_roles = this.current_user.roles
    axios.get('/api/roles').then(function (response) {
      self.available_roles = response.data.sort()
    }).catch(err => console.log(err))
  },
  methods: {
    memberEdit () {
      var self = this
      self.updating = true
      if (self.current_user.uuid !== undefined) {
        axios.put(
          `/api/admins/${self.uuid}/members/${this.current_user.uuid}`,
          this.current_user,
          { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.updating = false
          self.current_user = response.data
          self.notifyOK()
        }).catch(function (error) {
          self.updating = false
          self.notifyNOK()
          console.log(error)
        })
      } else {
        axios.post(
          `/api/admins/${this.uuid}/members`,
          this.current_user,
          { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.updating = false
          self.notifyOK()
        }).catch(function (error) {
          self.updating = false
          self.notifyNOK()
          console.log(error)
        })
      }
    },
    notifyOK () {
      const notification = {
        template: `<span>The member was successfully added ! He or she will receive an email with infos to connect.</span>`
      }
      this.$notifications.notify({
        component: notification,
        icon: 'nc-icon nc-check-2',
        type: 'success',
        timeout: 10000,
        showClose: false
      })
    },
    notifyNOK () {
      const notification = {
        template: `<span>There was an error during the member registration.</span>`
      }
      this.$notifications.notify({
        component: notification,
        icon: 'nc-icon nc-simple-remove',
        type: 'danger',
        timeout: null,
        showClose: false
      })
    }

  }
}
</script>
<style>
</style>
