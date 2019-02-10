const axios = require('axios')

export var practiceMixin = {
  methods: {
    deletePractice: function (practice) {
      var self = this
      var options = { year: 'numeric', month: '2-digit', day: '2-digit' }
      var date = new Date(practice.startDate * 1000)
      var startDate = Intl.DateTimeFormat('fr-FR', options).format(date)
      return new Promise((resolve, reject) => {
        this.$dialog
        .confirm(this.$t('practices.confirm_delete') + ' ' + practice.name + ' du ' + startDate + ' ?',
          { okText: this.$t('practices.ok_delete'), cancelText: this.$t('practices.cancel_delete') })
          .then(function () {
            axios.delete(
              `api/admins/${self.uuid}/events/${practice.uuid}`,
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
