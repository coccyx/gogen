<template>
  <div>
    <v-alert error
             v-model="backendError">{{ backendErrorMsg }}</v-alert>
    <v-spacer class="mt-5" />
    <v-row>
      <v-col xs4="xs4">
        <single-value header="Total Events"
                      v-bind:value="totalEvents" />
      </v-col>
      <v-col xs4="xs4">
        <single-value header="Total Bytes"
                      v-bind:value="totalBytes" />
      </v-col>
      <v-col xs4="xs4">
        <single-value header="Total Sources"
                      v-bind:value="totalSources" />
      </v-col>
    </v-row>
    <v-spacer class="mt-2" />
    <v-row>
      <v-col xs6="xs6">
        <single-value header="Generator Queue Depth"
                      v-bind:value="currentGeneratorQueue" />
      </v-col>
      <v-col xs6="xs6">
        <single-value header="Output Queue Depth"
                      v-bind:value="currentOutputQueue" />
      </v-col>
    </v-row>
    <v-spacer class="mt-2" />
    <v-row>
      <v-col xs6="xs6">
        <single-value header="Events Per Second"
                      v-bind:value="eventsPerSec" />
      </v-col>
      <v-col xs6="xs6">
        <single-value header="Bytes Per Second"
                      v-bind:value="bytesPerSec" />
      </v-col>
    </v-row>
  </div>
</template>

<script>
import { mapState } from 'vuex'
const prettyBytes = require('pretty-bytes');
const metric_suffix = require('metric-suffix');

export default {
  // Used for sending data on what this view is back to the main App component
  mounted() {
    this.$emit('view', this.meta())
  },
  preFetch() {
    return this.methods.meta()
  },
  computed: mapState({
    totalEvents: state => metric_suffix(state.totalEvents),
    totalBytes: state => prettyBytes(state.totalBytes),
    totalSources: state => state.totalSources.toString(),
    eventsPerSec: state => metric_suffix(state.eventsPerSec),
    bytesPerSec: state => prettyBytes(state.bytesPerSec),
    currentGeneratorQueue: state => state.currentGeneratorQueue.toString(),
    currentOutputQueue: state => state.currentOutputQueue.toString(),
    backendError: 'backendError',
    backendErrorMsg: 'backendErrorMsg'
  }),
  methods: {
    meta() {
      return {
        title: 'Realtime Gogen Stats',
        h1: 'Stats'
      }
    }
  }
}
</script>