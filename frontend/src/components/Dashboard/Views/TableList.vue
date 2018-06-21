<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-12">
          <card>
            <template slot="header">
              <h4 class="card-title">Striped Table with Hover</h4>
              <p class="card-category">Here is a subtitle for this table</p>
            </template>
            <div class="table-responsive">
              <l-table class="table-hover table-striped"
                       :columns="table.columns"
                       :data="table.data">
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
  const tableColumns = ['uuid', 'name', 'startDate', 'endDate']
  var tableTest = [{
    uuid: 1,
    name: 'Dakota Rice',
    salary: '$36.738',
    country: 'Niger',
    city: 'Oud-Turnhout'
  },
  {
    id: 2,
    name: 'Minerva Hooper',
    salary: '$23,789',
    country: 'Curaçao',
    city: 'Sinaai-Waas'
  }]
  export default {
    components: {
      LTable,
      Card
    },
    data () {
      var table = {
        columns: ['uuid', 'name', 'startDate', 'endDate'],
        data: tableTest
      }
      axios.get("http://localhost:8080/events")
          .then(function(response) {
            table.data = response.data
            console.log(self.data)
          })
      return {
        table
      }
    }
  }
</script>
<style>
</style>
