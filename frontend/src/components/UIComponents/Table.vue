<template>
  <table class="table">
    <thead>
      <tr>
        <slot name="columns">
          <th v-for="column in columns">{{capitalizeFirstLetter(column)}}</th>
        </slot>
      </tr>
    </thead>
    <tbody>
    <tr v-for="(item, index) in data" v-bind:style="styles[index]">
      <slot :row="item">
        <td v-for="column in columns" v-if="hasValue(item, column)">{{itemValue(item, column)}}</td>
        <td v-if="hasValue(item, 'uuid')" style="display:none">{{itemValue(item, 'uuid')}}</td>
      </slot>
    </tr>
    </tbody>
  </table>
</template>
<script>
  export default {
    name: 'l-table',
    props: {
      columns: Array,
      data: Array,
      styles: Array
    },
    methods: {
      hasValue (item, column) {
        return item[column] !== 'undefined'
      },
      itemValue (item, column) {
        return item[column]
      },
      capitalizeFirstLetter (string) {
        return string.charAt(0).toUpperCase() + string.slice(1)
      }
    }
  }
</script>
<style>
</style>
