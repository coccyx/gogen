<template>
  <div>
    <v-spacer class="mt-5" />
    <v-row>
      <v-card class="primary">
        <v-card-text>
          This is the editor!
        </v-card-text>
      </v-card>
    </v-row>
  </div>
</template>

<script>
export default {
  mounted() {
    this.$emit('view', this.meta())
  },
  preFetch() {
    return this.methods.meta()
  },
  methods: {
    meta() {
      return {
        title: 'Edit Gogen Configs',
        h1: 'Editor'
      }
    }
  }
}
</script>