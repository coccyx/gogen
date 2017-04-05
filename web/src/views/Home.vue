<template>
  <div>
    Hello World!
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
        title: 'Home of Gogen!',
        h1: 'Gogen'
      }
    }
  }
}
</script>