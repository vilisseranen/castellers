export var notificationMixin = {
  methods: {
    notifyOK (text) {
      const notification = {
        template: '<span>' + text + '</span>'
      }
      this.$notifications.notify({
        component: notification,
        icon: 'nc-icon nc-check-2',
        type: 'success',
        timeout: 10000,
        showClose: false
      })
    },
    notifyNOK (text) {
      const notification = {
        template: '<span>' + text + '</span>'
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
