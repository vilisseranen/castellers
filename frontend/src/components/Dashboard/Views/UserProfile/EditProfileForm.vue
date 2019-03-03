<template>
  <card>
    <h4 slot="header" class="card-title">{{ $t('members.' + actionLabel) }}</h4>
    <form>
      <div class="row">
        <div class="col-md-4">
          <fg-input type="text"
                    label="ID"
                    :disabled="true"
                    v-model="current_user.uuid">
          </fg-input>
        </div>
        <div class="col-md-4">
        <fg-input :label="$t('members.type')" type="radio" required="true">
          <form slot="input">
              <PrettyRadio class="p-default p-curve" name="type" color="primary-o" value="member" v-model="current_user.type">{{ $t('members.type_member') }}</PrettyRadio>
              <PrettyRadio class="p-default p-curve" name="type" color="success-o" value="admin" v-model="current_user.type">{{ $t('members.type_admin') }}</PrettyRadio>
          </form>
        </fg-input>
        </div>
        <div class="col-md-4">
        <fg-input :label="$t('members.language')" type="radio" required="true">
          <form slot="input">
              <PrettyRadio class="p-default p-curve" name="type" value="fr" v-model="current_user.language">{{ $t('members.lang_fr') }}</PrettyRadio>
              <PrettyRadio class="p-default p-curve" name="type" value="en" v-model="current_user.language">{{ $t('members.lang_en') }}</PrettyRadio>
              <PrettyRadio class="p-default p-curve" name="type" value="cat" v-model="current_user.language">{{ $t('members.lang_cat') }}</PrettyRadio>
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
        <div class="col-md-3">
          <fg-input type="text"
                    :label="$t('members.height')"
                    :placeholder="heightExemple()"
                    v-model.lazy="heightDisplayed">
          </fg-input>
        </div>
        <div class="col-md-3">
        <fg-input :label="$t('members.units')" type="radio">
          <form slot="input">
              <PrettyRadio class="p-default p-curve" name="height" color="primary-o" value="cm" v-model="height_unit">{{ $t('members.cm') }}</PrettyRadio>
              <PrettyRadio class="p-default p-curve" name="height" color="success-o" value="ft" v-model="height_unit">{{ $t('members.ft') }}</PrettyRadio>
          </form>
        </fg-input>
        </div>
        <div class="col-md-3">
          <fg-input type="text"
                    :label="$t('members.weight')"
                    :placeholder="weightExemple()"
                    v-model="weightDisplayed">
          </fg-input>
        </div>
        <div class="col-md-3">
        <fg-input :label="$t('members.units')" type="radio">
          <form slot="input">
              <PrettyRadio class="p-default p-curve" name="weight" color="primary-o" value="kg" v-model="weight_unit">{{ $t('members.kg') }}</PrettyRadio>
              <PrettyRadio class="p-default p-curve" name="weight" color="success-o" value="lb" v-model="weight_unit">{{ $t('members.lb') }}</PrettyRadio>
          </form>
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
      <div class="row">
        <div class="col-md-2">
        <slot name="delete-button">
          <button type="submit" class="btn btn-danger btn-fill float-left" @click.prevent="memberDelete" v-if="current_user.uuid">
            {{ $t('members.delete_button') }}
          </button>
        </slot>
        </div>
        <div class="col-md-10">
        <slot name="update-button">
          <button slot="update_button" type="submit" class="btn btn-info btn-fill float-right" @click.prevent="memberEdit">
            {{ $t('members.' + actionLabel + '_button') }}
          </button>
        </slot>
        <div style="width:10px; height: 1px; float: right;"></div>
        <slot name="email-button">
          <button type="submit" style="text-align: center" class="btn btn-warning btn-fill float-right" @click.prevent="resendEmail" v-if="current_user.uuid && this.type === 'admin'">
            {{ $t('members.email_button') }}
          </button>
        </slot>
        </div>
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
import {notificationMixin} from 'src/components/mixins/notifications.js'

export default {
  mixins: [notificationMixin],
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
      get: function () {
        return this.user
      },
      set: function (newUuid) {
        this.current_user.uuid = newUuid
      }
    },
    heightDisplayed: {
      get: function () {
        var height
        if (this.height_unit === 'ft') {
          var inches = Math.round((parseInt(this.current_user.height) / 2.54) % 12)
          var feet = Math.floor((parseInt(this.current_user.height) / 2.54) / 12)
          if (inches === 12) {
            feet += 1
            inches = ''
          }
          if (!isNaN(inches) && !isNaN(feet)) {
            height = feet + "'" + inches
            return height
          }
        } else {
          height = parseFloat(this.current_user.height)
        }
        if (!isNaN(height)) {
          return Math.round(height)
        }
      },
      set: function (height) {
        var heightToSave
        if (this.height_unit === 'ft') {
          var heightParsed = height.split("'", 2)
          var feet = parseFloat(heightParsed[0])
          var inches = parseFloat(heightParsed[1])
          if (isNaN(inches)) {
            inches = 0
          }
          if (!isNaN(feet)) {
            heightToSave = feet * 2.54 * 12 + inches * 2.54
            this.current_user.height = heightToSave.toFixed(2)
          } else {
          }
        } else {
          heightToSave = parseFloat(height)
          if (!isNaN(heightToSave)) {
            this.current_user.height = heightToSave.toFixed(2)
          }
        }
        if (isNaN(heightToSave)) {
          this.current_user.height = ''
        }
      }
    },
    weightDisplayed: {
      get: function () {
        var weight
        if (this.weight_unit === 'lb') {
          weight = parseFloat(this.current_user.weight) * 2.205
        } else {
          weight = parseFloat(this.current_user.weight)
        }
        if (weight !== 'undefined' && !isNaN(weight)) {
          return Math.round(weight)
        }
      },
      set: function (weight) {
        var weightToSave = parseFloat(weight)
        if (!isNaN(weightToSave)) {
          if (this.weight_unit === 'lb') {
            weightToSave /= 2.205
            if (!isNaN(weightToSave)) {
              this.current_user.weight = weightToSave.toFixed(2)
            }
          } else {
            this.current_user.weight = weightToSave.toFixed(2)
          }
        }
      }
    }
  },
  data () {
    return {
      updating: false,
      available_roles: [],
      height_unit: 'cm',
      weight_unit: 'kg'
    }
  },
  mounted () {
    var self = this
    axios.get('/api/roles').then(function (response) {
      self.available_roles = response.data.sort()
    }).catch(err => console.log(err))
  },
  methods: {
    heightExemple () {
      return this.height_unit === 'cm' ? this.$t('members.cm_exemple') : this.$t('members.ft_exemple')
    },
    weightExemple () {
      return this.weight_unit === 'kg' ? this.$t('members.kg_exemple') : this.$t('members.lb_exemple')
    },
    resendEmail () {
      var self = this
      self.updating = true
      axios.get(
          `/api/admins/${self.uuid}/members/${this.current_user.uuid}/registration`,
          { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.updating = false
          self.notifyOK(self.$t('members.notify_success'))
        }).catch(function (error) {
          self.updating = false
          self.notifyNOK(self.$t('members.notify_error'))
          console.log(error)
        })
    },
    memberEdit () {
      var self = this
      self.updating = true
      if (self.current_user.uuid !== undefined) {
        var url
        if (self.type === 'admin') {
          url = `/api/admins/${self.uuid}/members/${this.current_user.uuid}`
        } else {
          url = `/api/members/${this.current_user.uuid}`
        }
        axios.put(
          url,
          this.current_user,
          { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.updating = false
          self.notifyOK(self.$t('members.notify_success'))
          self.$emit('updateUser', response.data.uuid)
        }).catch(function (error) {
          self.updating = false
          self.notifyNOK(self.$t('members.notify_error'))
          console.log(error)
        })
      } else {
        axios.post(
          `/api/admins/${this.uuid}/members`,
          this.current_user,
          { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.updating = false
          self.notifyOK(self.$t('members.notify_success'))
          self.$emit('updateUser', response.data.uuid)
        }).catch(function (error) {
          self.updating = false
          self.notifyNOK(self.$t('members.notify_error'))
          console.log(error)
        })
      }
    },
    memberDelete () {
      this.$emit('deleteUser', this.current_user)
    }
  }
}
</script>
<style>
</style>
