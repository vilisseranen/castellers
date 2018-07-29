<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-12">
          <card>
            <template slot="header">
              <h4 class="card-title">Members</h4>
              <p class="text-right card-category" v-on:click="addMember">Click here to add a new Member</p>
            </template>
            <div class="table-responsive"> 
              <l-table class="table-hover table-striped"
                       :columns="table.columns"
                       :data="table.data">
                <template slot="columns"></template>
                <template slot-scope="{row}">
                  <td>{{row.firstName}}</td>
                  <td>{{row.lastName}}</td>
                  <td>{{row.roles}}</td>
                  <td>{{row.extra}}</td>
                  <td>{{row.type}}</td>
                  <td class="td-actions text-right" style="width: 40px">
                    <button type="button" class="btn-simple btn btn-xs btn-sucess" v-tooltip.top-center="editMember"
                            v-on:click="editMemberUuid(row.uuid)">
                      <i class="fa fa-edit"></i>
                    </button>
                    <button type="button" class="btn-simple btn btn-xs btn-sucess" v-tooltip.top-center="removeMember"
                            v-on:click="deleteMemberUuid(row.uuid)">
                      <i class=" 	fa fa-remove"></i>
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
    components: {
      LTable,
      Card
    },
    computed: {
      ...mapGetters(['uuid', 'code', 'type'])
    },
    data () {
      var table = {
        columns: ['firstName', 'lastName', 'roles', 'extra', 'type', 'actions'],
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
