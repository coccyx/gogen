<template>
  <div>
    <v-spacer class="mt-5" />
    <v-row>
      <v-card class="accent">
        <v-card-text>
          This is the stats view!
        </v-card-text>
      </v-card>
    </v-row>
  </div>
</template>

<script>
export default {
  // Used for sending data on what this view is back to the main App component
  mounted() {
    this.$emit('view', this.meta())
  },
  preFetch() {
    return this.methods.meta()
  },
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