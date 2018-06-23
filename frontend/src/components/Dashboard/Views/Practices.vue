<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-12">
          <card>
            <template slot="header">
              <h4 class="card-title">Upcoming practices</h4>
              <p class="card-category">You can find here a list of upcoming practices</p>
            </template>
            <div class="table-responsive"> 
              <l-table class="table-hover table-striped"
                       :columns="table.columns"
                       :data="table.data">
                <template slot="columns"></template>
                <template slot-scope="{row}">
                  <td>{{row.name}}</td>
                  <td>{{row.date}}</td>
                  <td>{{row.start}}</td>
                  <td>{{row.end}}</td>
                  <td class="td-actions text-right" style="width: 20px">
                    <button type="button" class="btn-simple btn btn-xs btn-sucess" v-tooltip.top-center="participateYes"
                            v-on:click="buttonClick('participate', row.member_uuid, row.name, row.uuid)">
                      <i class="fa fa-thumbs-o-up"></i>
                    </button>
                    <button type="button" class="btn-simple btn btn-xs btn-danger" v-tooltip.top-center="participateNo" v-on:click="buttonClick()">
                      <i class="fa fa-thumbs-down"></i>
                    </button>
                    <button type="button" class="btn-simple btn btn-xs btn-primary" v-tooltip.top-center="participateMaybe" v-on:click="buttonClick()">
                      <i class="fa fa-question-circle-o"></i>
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
<script>
  import LTable from 'src/components/UIComponents/Table.vue'
  import Card from 'src/components/UIComponents/Cards/Card.vue'
  import axios from 'axios'
  // const tableColumns = ['name', 'date', 'start', 'end', 'answer']
  const tableColumns = []
  export default {
    components: {
      LTable,
      Card
    },
    data () {
      var self = this
      var table = {
        columns: tableColumns,
        data: []
      }
      axios.get('http://localhost:8080/events')
          .then(function (response) {
            table.data = response.data
            for (var i = 0; i < table.data.length; i++) {
              table.data[i]['date'] = self.extractDate(table.data[i]['startDate'])
              table.data[i]['start'] = self.extractTime(table.data[i]['startDate'])
              table.data[i]['end'] = self.extractTime(table.data[i]['endDate'])
            }
          }).catch(err => console.log(err))
      return {
        table,
        participateYes: 'Participate',
        participateNo: 'Do not participate',
        participateMaybe: 'May or may not participate'
      }
    },
    methods: {
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
      buttonClick (participation, memberId, eventName, eventId) {
        console.log('I (' + memberId + ') will ' + participation + ' to the event: ' + eventName + ' (' + eventId + ')')
      }
    }
  }
</script>
<style>
</style>
