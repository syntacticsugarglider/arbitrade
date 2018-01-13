<template>
  <div id="app">
    <div class="nav">
      <div class="logo">ARBITRADE</div>
    </div>
    <router-view :symbols="symbols" />
  </div>
</template>

<script>
import io from 'socket.io-client';
export default {
  name: 'app',
  data() {
    return {
      symbols: []
    }
  },
  mounted() {
    const socket = io(`http://${window.location.hostname}:3001`);
    socket.on('symbols', (data) => {
      this.symbols = data;
    })
    socket.on('specSymbol', (data) => {
      this.symbols[data[0]] = data[1];
    })
  }
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
body {
  margin: 0;
  background: black;
  color: white;
}
.nav {
  height: 60px;
  position: fixed;
  top: 0;
  width: 100%;
  background-color: #1a1a1a;
  display: flex;
  justify-content: center;
  flex-flow: row nowrap;
  align-items: center;
  box-shadow: 0 3px 6px rgba(0,0,0,0.16), 0 3px 6px rgba(0,0,0,0.23);
}
.logo {
  left: 20px;
  position: fixed;
  color: #FDD835;
  font-weight: 800;
  font-size: 2rem;
}
.item {
  padding: 10px;
  background: rgba(253,216,53,0.8);
  opacity: 1;
  transition: 0.2s background-color ease;
  font-weight: 800;
  border-radius: 2px;
  color: black;
  cursor: pointer;
}
.item:not(:last-child) {
  margin-right: 10px;
}
.item:hover {
  background: rgba(253,216,53,0.9);
}
</style>
