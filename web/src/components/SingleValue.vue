<template>
  <v-card>
    <v-card-row height="75px"
                class="deep-orange darken-2">
      <v-card-text class="text-xs-center display-1 grey--text text--lighten-3">{{ header }}</v-card-text>
    </v-card-row>
    <v-card-row height="125px"
                class="deep-orange darken-4">
      <v-card-text class="text-xs-center display-3 grey--text text--lighten-4">{{ value }}</v-card-text>
    </v-card-row>
  </v-card>
</template>

<script>
export default {
  name: 'single-value',

  props: {
    header: String,
    value: String,
  }
}
</script>