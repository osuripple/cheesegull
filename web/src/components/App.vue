<template lang="pug">
#app
  search-bar
</template>

<script>
  import SearchBar from './SearchBar.vue'
  const components = {
    'search-bar': SearchBar,
  }
  export default {
    components,
  }
</script>

<style>
  @import url('https://fonts.googleapis.com/css?family=Lato');
  body {
    font-family: "Lato", sans-serif;
  }
</style>

<style src="../../node_modules/normalize.css/normalize.css"></style>
