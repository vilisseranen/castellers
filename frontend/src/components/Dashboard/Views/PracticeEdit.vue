<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-md-12">
          <edit-practice-form :event="event" :updating="updating" v-on:updatePractice="loadEvent">
            <template slot="message">
              <span></span>
            </template>
          </edit-practice-form>
        </div>
      </div>
    </div>
    <card>
      <template slot="header">
        <h4 class="card-title">{{ $t('practices.participants') }}</h4>
        <p class="text-right">
          <toggle-button
                  v-model="displayAllMembers"
                  color="#82C7EB"
                  :sync="true"/>
          {{ $t('practices.display_all_members') }}
        </p>
      </template>
      <div class="table-responsive"> 
        <l-table class="table-hover table-striped"
                  :columns="columns.map(x => $t('practices.' + x))"
                  :data="table.data"
                  :styles="table.styles">
          <template slot="columns"></template>
          <template slot-scope="{row}">
            <td v-if="row.participation === 'yes' || displayAllMembers === true">{{row.firstName}} {{row.lastName}}</td>
            <td v-if="row.participation === 'yes' || displayAllMembers === true">{{row.roles.join(", ")}}</td>
            <td v-if="row.participation === 'yes' || displayAllMembers === true">{{ $t('practices.participation_' + row.participation) }}</td>
          </template>
        </l-table>
      </div>
      <p slot="footer">
        {{ $t('practices.totals_participants')}}:
        {{ countParticipants('yes') }} {{ $t('practices.participation_yes').toLowerCase()}},
        {{ countParticipants('maybe') }} {{ $t('practices.participation_maybe').toLowerCase()}},
        {{ countParticipants('no') }} {{ $t('practices.participation_no').toLowerCase()}},
        {{ countParticipants('') }} {{ $t('practices.participation_').toLowerCase()}}
      </p>
    </card>
  </div>
</template>

<i18n src='assets/translations/practices.json'></i18n>

<script>
import EditPracticeForm from './EditPracticeForm.vue'
import Card from 'src/components/UIComponents/Cards/Card.vue'
import LTable from 'src/components/UIComponents/Table.vue'
import axios from 'axios'
import {mapGetters} from 'vuex'

export default {
  components: {
    EditPracticeForm,
    LTable,
    Card
  },
  data () {
    var now = new Date(Date.now())
    var table = {
      data: []
    }
    return {
      event: {
        'startDate': Math.trunc(now.valueOf() / 1000),
        'endDate': Math.trunc(now.valueOf() / 1000),
        'recurring': {
          'interval': '1w',
          'until': 0
        }
      }, // defaults are set here
      updating: false,
      table,
      displayAllMembers: false
    }
  },
  mounted () {
    this.loadEvent(this.$route.params.uuid)
    this.listParticipants(this.$route.params.uuid)
  },
  computed: {
    ...mapGetters(['uuid', 'code', 'type']),
    columns: function () {
      return ['participant_name', 'roles', 'participation']
    }
  },
  methods: {
    countParticipants (participation) {
      return this.table.data.filter(participant => participant.participation === participation).length
    },
    loadEvent (uuid) {
      if (uuid) {
        var self = this
        var url = `/api/events/${uuid}`
        axios.get(
          url, { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.event = response.data
          self.$router.push({path: `/practiceEdit/${self.event.uuid}`})
        }).catch(err => console.log(err))
      }
    },
    listParticipants (uuid) {
      if (uuid) {
        var self = this
        var url = `/api/admins/${this.uuid}/events/${uuid}/members`
        axios.get(
          url, { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.table.data = response.data
          for (var i = 0; i < self.table.data.length; i++) {
            if (self.table.data[i]['participation'] === 'yes') {
              self.table.data[i]['style'] = { background: 'rgba(174, 224, 127, 0.25)' }
            } else if (self.table.data[i]['participation'] === 'no') {
              self.table.data[i]['style'] = { background: 'rgba(232, 78, 78, 0.25)' }
            } else if (self.table.data[i]['participation'] === 'maybe') {
              self.table.data[i]['style'] = { background: 'rgba(232, 178, 8, 0.25)' }
            } else {
              self.table.data[i]['style'] = { background: 'rgba(7, 124, 232, 0.25)' }
            }
          }
        }).catch(err => console.log(err))
      }
    }
  }
}
</script>
<style>
</style>
