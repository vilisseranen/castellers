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
  </div>
</template>

<i18n src='assets/translations/practices.json'></i18n>

<script>
import EditPracticeForm from './EditPracticeForm.vue'
import axios from 'axios'
import {mapGetters} from 'vuex'

export default {
  components: {
    EditPracticeForm
  },
  data () {
    var now = new Date(Date.now())
    return {
      event: {
        'startDate': Math.trunc(now.valueOf() / 1000),
        'endDate': Math.trunc(now.valueOf() / 1000),
        'recurring': {
          'interval': '1w',
          'until': 0
        }
      }, // defaults are set here
      updating: false
    }
  },
  mounted () {
    this.loadEvent(this.$route.params.uuid)
  },
  computed: {
    ...mapGetters(['uuid', 'code', 'type'])
  },
  methods: {
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
    }
  }
}
</script>
<style>
</style>
