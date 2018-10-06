<template>
  <div class="content">
    <div class="container-fluid">
      <div class="row">
        <div class="col-12">
          <card>
            <template slot="header">
              <h4 class="card-title">{{ $t('members.title') }}</h4>
              <p class="text-right card-category" v-on:click="addMember">{{ $t('members.add_a_member') }} <i class="fa fa-user-plus member-icon"></i></p>
            </template>
            <div class="table-responsive"> 
              <l-table class="table-hover table-striped"
                       :columns="columns.map(x => $t('members.' + x))"
                       :data="table.data">
                <template slot="columns"></template>
                <template slot-scope="{row}">
                  <td>{{row.firstName}}</td>
                  <td>{{row.lastName}}</td>
                  <td>{{row.roles.join(", ")}}</td>
                  <td>{{row.extra}}</td>
                  <td>{{row.type}}</td>
                  <td class="td-actions text-right" style="width: 40px">
                    <button type="button" class="btn-simple btn btn-xs btn-info" v-tooltip.top-center="$t('members.edit')"
                            v-on:click="editMemberUuid(row.uuid)">
                      <i class="fa fa-edit"></i>
                    </button>
                    <button type="button" class="btn-simple btn btn-xs btn-danger" v-tooltip.top-center="$t('members.remove')"
                            v-on:click="deleteMemberUuid(row.uuid, row.firstName, row.lastName)">
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

<i18n src='assets/translations/members.json'></i18n>

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
      ...mapGetters(['uuid', 'code', 'type']),
      columns: function () {
        return ['first_name', 'last_name', 'roles', 'extra', 'type', 'actions']
      }
    },
    data () {
      var table = {
        data: []
      }
      return {
        table
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
      deleteMemberUuid (memberUuid, memberFirstName, memberLastName) {
        var self = this
        this.$dialog
        .confirm(this.$i18n.t('members.confirm_delete') + ' ' + memberFirstName + ' ' + memberLastName + ' ?',
                 {okText: this.$i18n.t('members.ok_delete'), cancelText: this.$i18n.t('members.cancel_delete')})
        .then(function(dialog) {
          axios.delete(
            `api/admins/${self.uuid}/members/${memberUuid}`,
            { headers: { 'X-Member-Code': self.code } }
          ).then(function (response) {
            self.listMembers()
          }).catch(err => console.log(err))
        });
      }
    }
  }
</script>
<style>
</style>
