<template>
  <card>
    <h4 slot="header" class="card-title">{{ $t('practices.' + actionLabel) }}</h4>
    <form>
      <div class="row">
        <div class="col-md-4">
          <fg-input type="text"
                    label="ID"
                    :disabled="true"
                    v-model="current_event.uuid">
          </fg-input>
        </div>
        <div class="col-md-8">
          <fg-input type="text"
                    :label="$t('practices.name')"
                    :placeholder="$t('practices.name_description')"
                    v-model="current_event.name"
                    required="true">
          </fg-input>
        </div>
      </div>
      <div class="row">
        <div class="col-md-6">
          <fg-input type="number"
                    :label="$t('practices.start')"
                    required="true">
            <template slot="input">
              <VueCtkDateTimePicker minuteInterval=15 v-model="startDateForCalendar" />
            </template>
          </fg-input>          
        </div>
        <div class="col-md-6">
          <fg-input type="number"
                    :label="$t('practices.end')"
                    required="true">
            <template slot="input">
              <VueCtkDateTimePicker minuteInterval=15 v-model="endDateForCalendar" />
            </template>
          </fg-input>
        </div>
      </div>
      <div class="row">
        <div class="col-md-12">
        <slot name="update-button">
          <button slot="update_button" type="submit" class="btn btn-info btn-fill float-right" @click.prevent="practiceEdit">
            {{ $t('practices.' + actionLabel + '_button') }}
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

<i18n src='assets/translations/practices.json'></i18n>
<i18n src='assets/translations/general.json'></i18n>

<script>
import Card from 'src/components/UIComponents/Cards/Card.vue'
import VueCtkDateTimePicker from 'vue-ctk-date-time-picker'
import axios from 'axios'
import {mapGetters} from 'vuex'

import 'vue-ctk-date-time-picker/dist/vue-ctk-date-time-picker.css'

export default {
  components: {
    Card,
    VueCtkDateTimePicker
  },
  name: 'edit-practice-form',
  props: {
    event: Object
  },
  computed: {
    ...mapGetters(['uuid', 'code', 'type', 'language']),
    actionLabel: function () {
      return this.event.uuid ? 'update' : 'create'
    },
    current_event: {
      get: function () {
        return this.event
      },
      set: function (newUuid) {
        this.current_event.uuid = newUuid
      }
    },
    startDateForCalendar: {
      get: function () {
        return this.dateToCalendar(this.current_event.startDate)
      },
      set: function (newDate) {
        this.current_event.startDate = this.dateFromCalendar(newDate)
      }
    },
    endDateForCalendar: {
      get: function () {
        return this.dateToCalendar(this.current_event.endDate)
      },
      set: function (newDate) {
        this.current_event.endDate = this.dateFromCalendar(newDate)
      }
    }
  },
  data () {
    return {
      updating: false
    }
  },
  methods: {
    dateFromCalendar (dateToConvert) {
      if (dateToConvert) {
        var date = new Date(dateToConvert)
        return Math.trunc(date.getTime() / 1000)
      } else {
        return 0
      }
    },
    dateToCalendar (dateToConvert) {
      if (dateToConvert) {
        var date = new Date(dateToConvert * 1000)
        return date.getFullYear() + '-' + (date.getMonth() + 1) + '-' + date.getDate() +
             ' ' + date.getHours() + ' ' + date.getMinutes()
      } else {
        return ''
      }
    },
    practiceEdit () {
      var self = this
      self.updating = true
      if (self.current_event.uuid !== undefined) {
        var url = `/api/admins/${self.uuid}/events/${this.current_event.uuid}`
        axios.put(
          url,
          this.current_event,
          { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.updating = false
          self.notifyOK()
          self.$emit('updatePractice', response.data.uuid)
        }).catch(function (error) {
          self.updating = false
          self.notifyNOK()
          console.log(error)
        })
      } else {
        axios.post(
          `/api/admins/${this.uuid}/events`,
          this.current_event,
          { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.updating = false
          self.notifyOK()
          self.$emit('updatePractice', response.data.uuid)
        }).catch(function (error) {
          self.updating = false
          self.notifyNOK()
          console.log(error)
        })
      }
    },
    notifyOK () {
      const notification = {
        template: '<span>' + this.$i18n.t('practices.notify_success') + '</span>'
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
        template: '<span>' + this.$i18n.t('practices.notify_error') + '</span>'
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
