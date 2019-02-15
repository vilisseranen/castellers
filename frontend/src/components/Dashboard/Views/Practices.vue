<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-12">
          <card>
            <template slot="header">
              <h4 class="card-title">{{ $t('practices.title') }}</h4>
              <p class="text-right card-category" v-if="type == 'admin'" v-on:click="addPractice">{{ $t('practices.create') }} <i class="nc-icon nc-notification-70"></i></p>
            </template>
            <div class="table-responsive"> 
              <l-table class="table-hover table-striped"
                       :columns="columns.map(x => $t('practices.' + x))"
                       :data="table.data"
                       :styles="table.styles">
                <template slot="columns"></template>
                <template slot-scope="{row}">
                  <td>{{row.name}}</td>
                  <td>{{row.date}}</td>
                  <td>{{row.start}}</td>
                  <td>{{row.end}}</td>
                  <td v-if="type == 'admin'" class="td-actions text-center" style="width: 40px">
                    <button type="button" class="btn-simple btn btn-xs btn-info" v-tooltip.top-center="$t('practices.edit')"
                            v-on:click="editPracticeUuid(row.uuid)">
                      <i class="fa fa-edit"></i>
                    </button>
                    <button type="button" class="btn-simple btn btn-xs btn-danger" v-tooltip.top-center="$t('practices.remove')"
                            v-on:click="removePractice(row)">
                      <i class="fa fa-remove"></i>
                    </button>
                  </td>
                  <td v-if="uuid" class="td-actions text-right" style="width: 40px">
                    <button type="button" class="btn-simple btn btn-xs btn-sucess" v-tooltip.top-center="$t('practices.participate_yes')"
                            v-on:click="participation(row.uuid, 'yes')">
                      <i class="fa fa-thumbs-o-up"></i>
                    </button>
                    <button type="button" class="btn-simple btn btn-xs btn-danger" v-tooltip.top-center="$t('practices.participate_no')"
                            v-on:click="participation(row.uuid, 'no')">
                      <i class="fa fa-thumbs-down"></i>
                    </button>
                  </td>
                </template>
              </l-table>
            </div>
          </card>

        </div>

      </div>
    </div>
  </div>
</template>

<i18n src='assets/translations/practices.json'></i18n>

<script>
  import LTable from 'src/components/UIComponents/Table.vue'
  import Card from 'src/components/UIComponents/Cards/Card.vue'
  import axios from 'axios'
  import {mapGetters} from 'vuex'
  import {practiceMixin} from 'src/components/mixins/practices.js'
  import {notificationMixin} from 'src/components/mixins/notifications.js'

  export default {
    mixins: [practiceMixin, notificationMixin],
    components: {
      LTable,
      Card
    },
    computed: {
      ...mapGetters(['uuid', 'code', 'type']),
      columns: function () {
        var baseColumns = ['name', 'date', 'start', 'end']
        if (this.type === 'admin') {
          baseColumns.push('actions')
        }
        if (this.uuid) {
          baseColumns.push('participation')
        }
        return baseColumns
      }
    },
    mounted () {
      this.listPractices()
    },
    data () {
      var table = {
        data: [],
        styles: []
      }
      return {
        table
      }
    },
    methods: {
      listPractices () {
        var self = this
        axios.get('/api/events')
          .then(function (response) {
            self.table.data = response.data
            for (var i = 0; i < self.table.data.length; i++) {
              self.table.data[i]['date'] = self.extractDate(self.table.data[i]['startDate'])
              self.table.data[i]['start'] = self.extractTime(self.table.data[i]['startDate'])
              self.table.data[i]['end'] = self.extractTime(self.table.data[i]['endDate'])
              if (self.uuid) {
                var styles = []
                axios.get(`/api/events/${self.table.data[i]['uuid']}/members/${self.uuid}`,
                { headers: { 'X-Member-Code': self.code } })
                .then(function (response) {
                  if (response.status === 200) {
                    if (response.data.answer === 'yes') {
                      styles.push({ background: 'rgba(174, 224, 127, 0.25)' }) // rgba(174, 224, 127, 0.3)
                    } else if (response.data.answer === 'no') {
                      styles.push({ background: 'rgba(232, 78, 78, 0.25)' }) // rgba(232, 78, 78, 0.3)
                    }
                  }
                })
                self.table.styles = styles
              }
            }
          }).catch(err => console.log(err))
      },
      extractDate (timestamp) {
        var options = { year: 'numeric', month: '2-digit', day: '2-digit' }
        var date = new Date(timestamp * 1000)
        return new Intl.DateTimeFormat('fr-FR', options).format(date)
      },
      extractTime (timestamp) {
        var options = { hour: '2-digit', minute: '2-digit' }
        var time = new Date(timestamp * 1000)
        return new Intl.DateTimeFormat('fr-FR', options).format(time)
      },
      participation (eventuuid, participation) {
        var self = this
        axios.post(
          `/api/events/${eventuuid}/members/${this.uuid}`,
          { 'answer': participation },
          { headers: { 'X-Member-Code': this.code } }
          ).then(function () {
            self.notifyOK(self.$t('practices.participation_ok'))
            self.listPractices()
          }).catch(function () {
            self.notifyNOK(self.$t('practices.participation_nok'))
          })
      },
      addPractice () {
        this.$router.push({name: 'PracticeAdd'})
      },
      editPracticeUuid (practiceUuid) {
        this.$router.push({path: `/practiceEdit/${practiceUuid}`})
      },
      removePractice (practice) {
        var self = this
        this.deletePractice(practice)
          .then(function () { self.listPractices() })
          .catch(function (error) { console.log(error) })
      }
    }
  }
</script>
<style>
</style>
