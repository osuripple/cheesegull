<template lang="pug">
input(type="text", autofocus, v-model="searched", placeholder="Search...
").search-bar
</template>

<script>
  export default {
    data: () => {
      return {
        searched: ""
      }
    }
  }
</script>

<style>
.search-bar {
  width: 100%;
  height: 60px;
  font-size: 40px;
  position: fixed;
  top: 0;
  border: 0;
  padding: 0 10px;
  margin: 0;
  outline: 0;
  background: #e2edff;
  font-family: "Lato", sans-serif;
}
</style>
