<template>
  <div id="app">
    <div class="nav">
      <div class="logo">ARBITRADE</div>
      <div @click="$router.push('/')" class="item">OPPORTUNITIES</div>
      <div @click="$router.push('/trackers')" v-if="hTracked.length>0" class="item">TRADING</div>
      <div @click="$router.push('/balances')" class="item">BALANCES</div>
    </div>
    <router-view :trackers="rtTrackers" :symbols="symbols" :track="track" :hTracked="hTracked" />
  </div>
</template>

<script>
import io from 'socket.io-client';
export default {
  name: 'app',
  data() {
    return {
      symbols: [],
      hTracked: [],
      socket: undefined,
      rtTrackers: {}
    }
  },
  mounted() {
    this.socket = io(`http://${window.location.hostname}:3001`);
    this.socket.on('symbols', (data) => {
      this.symbols = Object.assign({}, data);
    })
    this.socket.on('specSymbol', (data) => {
      this.symbols[data[0]] = data[1];
    })
    this.socket.on('hTracked', (data) => {
      this.hTracked = data;
    })
    this.socket.on('hTrackData', (data) => {
      this.rtTrackers[data.symbol] = Object.assign(this.rtTrackers[data.symbol] ? this.rtTrackers[data.symbol] : {}, data);
      this.rtTrackers = JSON.parse(JSON.stringify(this.rtTrackers));
      this.$emit('update','track');
    })
  },
  methods: {
    track: function (obj) {
      if (this.hTracked.indexOf(obj)===-1) {
        this.hTracked.push(obj);
        this.socket.emit('hTrackedP', obj);
      }
      else {
        this.hTracked.splice(this.hTracked.indexOf(obj),1);
        this.socket.emit('hTrackedM', obj);
      }
    }
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
  background-color: #1a1a1a;
  color: white;
}
.nav {
  height: 60px;
  position: fixed;
  top: 0;
  width: 100%;
  background-color: #1a1a1a;
  display: flex;
  z-index: 100;
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
  user-select: none;
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
* {
  box-sizing: border-box;
}
::-webkit-scrollbar {
    width: 0px;  /* remove scrollbar space */
    background: transparent;  /* optional: just make scrollbar invisible */
}
/* optional: show position indicator in red */
::-webkit-scrollbar-thumb {
    background: #FF0000;
}
</style>
