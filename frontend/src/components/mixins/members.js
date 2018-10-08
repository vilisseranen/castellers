const axios = require('axios')

export var memberMixin = {
  methods: {
    deleteUser: function (member) {
      var self = this
      return new Promise((resolve, reject) => {
        this.$dialog
        .confirm(this.$t('members.confirm_delete') + ' ' + member.firstName + ' ' + member.lastName + ' ?',
          { okText: this.$t('members.ok_delete'), cancelText: this.$t('members.cancel_delete') })
          .then(function () {
            axios.delete(
              `api/admins/${self.uuid}/members/${member.uuid}`,
              { headers: { 'X-Member-Code': self.code } }
            )
            .then(function () {
              resolve()
            })
            .catch(function (err) {
              reject(err)
            })
          })
          .catch(function (err) {
            reject(err)
          })
      })
    }
  }
}
