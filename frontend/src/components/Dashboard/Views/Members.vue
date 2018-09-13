<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-12">
          <card>
            <template slot="header">
              <h4 class="card-title">Members</h4>
              <p class="text-right card-category" v-on:click="addMember">{{ $t('test.test') }} <i class="fa fa-user-plus"></i></p>
            </template>
            <div class="table-responsive"> 
              <l-table class="table-hover table-striped"
                       :columns="columns"
                       :data="table.data">
                <template slot="columns"></template>
                <template slot-scope="{row}">
                  <td>{{row.firstName}}</td>
                  <td>{{row.lastName}}</td>
                  <td>{{row.roles.join(", ")}}</td>
                  <td>{{row.extra}}</td>
                  <td>{{row.type}}</td>
                  <td class="td-actions text-right" style="width: 40px">
                    <button type="button" class="btn-simple btn btn-xs btn-info" v-tooltip.top-center="editMember"
                            v-on:click="editMemberUuid(row.uuid)">
                      <i class="fa fa-edit"></i>
                    </button>
                    <button type="button" class="btn-simple btn btn-xs btn-danger" v-tooltip.top-center="removeMember"
                            v-on:click="deleteMemberUuid(row.uuid)">
                      <i class="fa fa-remove"></i>
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
  import {mapGetters} from 'vuex'

  export default {
    i18n: {
      messages: {
    fr: {
      test: {
        test: "Hello from component"
      }
    }
  }
    },
    components: {
      LTable,
      Card
    },
    computed: {
      ...mapGetters(['uuid', 'code', 'type']),
      columns: function () {
        return ['First name', 'Last name', 'Roles', 'Extra', 'Type', 'Actions']
      }
    },
    data () {
      var table = {
        data: []
      }
      return {
        table,
        editMember: 'Edit',
        removeMember: 'Remove'
      }
    },
    mounted () {
      this.listMembers()
    },
    methods: {
      listMembers () {
        var self = this
        axios.get(
          `/api/admins/${this.uuid}/members`,
          { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.table.data = response.data
        }).catch(err => console.log(err))
      },
      addMember () {
        this.$router.push('memberAdd')
      },
      editMemberUuid (memberUuid) {
        this.$router.push({path: `MemberEdit/${memberUuid}`})
      },
      deleteMemberUuid (memberUuid) {
        var self = this
        axios.delete(
          `api/admins/${this.uuid}/members/${memberUuid}`,
          { headers: { 'X-Member-Code': this.code } }
        ).then(function (response) {
          self.listMembers()
        }).catch(err => console.log(err))
      }
    }
  }
</script>
<style>
</style>
