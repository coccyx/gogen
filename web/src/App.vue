<template>
  <v-app left-fixed-sidebar>
    <header>
      <v-toolbar>
        <v-toolbar-side-icon class="hidden-md-and-up"
                             @click.native.stop="sidebar = !sidebar" />
        <v-toolbar-logo v-text="title" />
      </v-toolbar>
    </header>
    <main>
      <v-sidebar v-model="sidebar"
                 fixed
                 class="grey lighten-4">
        <div class="nav-header">
          <router-link to="/about/">
            <img src="~public/Gogen.png"
                 alt="Gogen Logo" />
          </router-link>
        </div>
        <v-list dense>
          <v-list-item v-for="item in menu"
                       :key="item.title">
            <v-list-tile :href="item.href"
                         ripple
                         router>
              <v-list-tile-avatar>
                <v-icon class="grey--text text--darken-3">{{ item.icon }}</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title class="grey--text text--darken-3"
                                   v-text="item.title" />
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item>
        </v-list>
      </v-sidebar>
      <v-content>
        <v-container>
          <router-view @view="meta"></router-view>
        </v-container>
      </v-content>
    </main>
  </v-app>
</template>



<script>
export default {
  data() {
    return {
      sidebar: true,
      menu: [
        { icon: "home", title: "Home", href: "/" },
        { icon: "mode_edit", title: "Editor", href: "/editor/" },
        { icon: "timeline", title: "Realtime Stats", href: "/stats/" }
      ],
      title: null
    }
  },
  methods: {
    meta(obj) {
      this.title = obj.h1
      this.$store.commit('gogen/TITLE', obj.h1)
    }
  }
}
</script>

<style lang="stylus">
  $theme := {
    primary: #b71c1c
    accent: #d50000
    secondary: #424242
    info: #0D47A1
    warning: #ffb300
    error: #B71C1C
    success: #2E7D32
  }

  @import '../node_modules/vuetify/src/stylus/main'
  @import './css/main.css'
</style>
