<template>
  <card>
    <h4 slot="header" class="card-title">{{ $t('members.' + actionLabel) }}</h4>
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
        <fg-input label="type" type="radio" required="true">
          <form slot="input">
              <PrettyRadio class="p-default p-curve" name="type" color="primary-o" value="member" v-model="current_user.type">{{ $t('members.type_member') }}</PrettyRadio>
              <PrettyRadio class="p-default p-curve" name="type" color="success-o" value="admin" v-model="current_user.type">{{ $t('members.type_admin') }}</PrettyRadio>
          </form>
        </fg-input>
        </div>
      </div>
      <div class="row">
        <div class="col-md-4">
          <fg-input type="text"
                    :label="$t('members.first_name')"
                    :placeholder="$t('members.first_name')"
                    v-model="current_user.firstName"
                    required="true">
          </fg-input>
        </div>
        <div class="col-md-4">
          <fg-input type="text"
                    :label="$t('members.last_name')"
                    :placeholder="$t('members.last_name')"
                    v-model="current_user.lastName"
                    required="true">
          </fg-input>
        </div>
        <div class="col-md-4">
          <fg-input type="email"
                    :label="$t('members.email')"
                    :placeholder="$t('members.email')"
                    v-model="current_user.email"
                    required="true">
          </fg-input>
        </div>
      </div>
      <div class="row">
        <div class="col-md-8">
          <fg-input type="text"
                    :label="$t('members.roles')">
            <template slot="input">
              <multiselect
                v-model="current_user.roles"
                :options="available_roles"
                :multiple="true"
                :placeholder="''"
                :closeOnSelect="false"
                :selectLabel="$t('multiselect.selectLabel')"
                :selectGroupLabel="$t('multiselect.selectGroupLabel')"
                :deselectLabel="$t('multiselect.deselectLabel')"
                :deselectGroupLabel="$t('multiselect.deselectGroupLabel')"
                :selectedLabel="$t('multiselect.selectedLabel')">
              </multiselect>
            </template>
          </fg-input>
        </div>
        <div class="col-md-4">
          <fg-input type="text"
                    :label="$t('members.extra')"
                    :placeholder="$t('members.extra')"
                    v-model="current_user.extra">
          </fg-input>
        </div>
      </div>
      <div slot="message" class="row">
        <div class="col-md-12">
          <div class="alert alert-success" v-if="current_user.activated === 1">
            <span>{{ $t('members.already_logged_in') }}</span>
          </div>
           <div class="alert alert-warning" v-if="current_user.activated === 0">
            <span>{{ $t('members.never_logged_in') }}</span>
          </div>
        </div>
      </div>
      <div class="text-center">
        <slot name="update-button">
          <button slot="update_button" type="submit" class="btn btn-info btn-fill float-right" @click.prevent="memberEdit">
            {{ $t('members.' + actionLabel + '_button') }}
          </button>
        </slot>
        <slot name="delete-button">
          <button type="submit" class="btn btn-danger btn-fill float-right" @click.prevent="memberDelete" v-if="current_user.uuid">
            {{ $t('members.delete_button') }}
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
    <div slot="footer" class="stats">
      <slot name="footer"><span class=required></span>{{ $t('general.required_fields') }}
      </slot>
    </div>
  </card>
</template>

<i18n src='assets/translations/members.json'></i18n>
<i18n src='assets/translations/multiselect.json'></i18n>
<i18n src='assets/translations/general.json'></i18n>

<script>
import Card from 'src/components/UIComponents/Cards/Card.vue'
import axios from 'axios'
import {mapGetters} from 'vuex'
import Multiselect from 'vue-multiselect'
import 'vue-multiselect/dist/vue-multiselect.min.css'
import PrettyRadio from 'pretty-checkbox-vue/radio'
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
      return this.user.uuid ? 'update' : 'create'
    },
    current_user: {
      get: function() {
        return this.user
      },
      set: function (newUuid) {
        this.current_user.uuid = newUuid
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
          self.current_user = response.data.uuid
          self.notifyOK()
        }).catch(function (error) {
          self.updating = false
          self.notifyNOK()
          console.log(error)
        })
      }
      this.$emit('updateUser', this.current_user.uuid)
    },
    memberDelete () {
      this.$emit('deleteUser', this.current_user)
    },
    notifyOK () {
      const notification = {
        template: '<span>' + this.$i18n.t('members.notify_success') + '</span>'
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
        template: '<span>' + this.$i18n.t('members.notify_error') + '</span>'
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
