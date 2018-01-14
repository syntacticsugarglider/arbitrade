<template>
  <div class="centerer">
  <div class="container">
    <div v-for="symbol in hTracked" class="tradeop">
      <div @click="$props.track(symbol)" class="untrack">UNTRACK</div>
      <div class="c">
        {{symbol}}
      </div>
      <div class="age">
        latest from API is {{((currentTime - trackers[symbol].timestamp)/1000).toFixed(1)}}s old
      </div>
      <div class="path">
        trade is from <strong>{{trackers[symbol].e1}}</strong> for <strong>{{trackers[symbol].ask1}}{{trackers[symbol].c2}}</strong>
      </div>
      to
      <div class="path">
        <strong>{{trackers[symbol].e2}}</strong> at <strong>{{trackers[symbol].bid2}}{{trackers[symbol].c2}}</strong>
      </div>
      for a profit of
      <div class="path">
        <div class="profit">{{(((trackers[symbol].bid2-trackers[symbol].ask1)/trackers[symbol].bid2)*100).toFixed(1)}}%</div>
      </div>
      <div class="tradebutton">
        <div contenteditable="true" class="tradeinput">100</div>$<div class="tradetext">TRADE</div>
      </div>
    </div>
  </div>
  </div>
</template>

<script>
export default {
  name: 'trading',
  props: ['hTracked', 'symbols', 'track', 'trackers'],
  data() {
    return {
      currentTime: Date.now()
    }
  },
  mounted() {
    setInterval(() => {
      this.currentTime = Date.now()
    },100)
  }
}
</script>

<style scoped>
.container {
  width: 100%;
  margin-top: 60px;
  display: flex;
  justify-content: flex-start;
  align-items: center;
  max-width: 960px;
  flex-flow: row wrap;
}
.centerer {
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1;
}
.tradeop {
  box-shadow: 0 3px 6px rgba(0,0,0,0.16), 0 3px 6px rgba(0,0,0,0.23);
  background: #FDD835;
  color: black;
  margin: 10px;
  padding: 20px;
  padding-top: 40px;
  width: 300px;
  display: flex;
  justify-content: flex-start;
  align-items: center;
  flex-flow: column nowrap;
  height: 400px;
  position: relative;
  border-radius: 2px;
}
.c {
  font-weight: 500;
  font-family: Futura;
  font-size: 2rem;
}
.untrack {
  position: absolute;
  top: 0;
  left: 0;
  background: black;
  color: #FDD835;
  border-radius: 2px;
  user-select: none;
  cursor: pointer;
  font-size: 0.7rem;
  padding: 10px;
  font-family: Futura;
}
.path {
  font-size: 0.7rem;
}
strong {
  font-size: 0.8rem;
}
.profit {
  font-size: 4rem;
  font-family: Futura;
}
.tradebutton {
  display: flex;
  justify-content: center;
  font-family: Futura;
  font-size: 1.5rem;
  border: 2px solid black;
  border-radius: 5px;
  align-items: center;
  flex-flow: row nowrap;
}
.tradeinput {
  outline: none;
  display: inline-block;
  margin: 10px;
  margin-right: 0;
}
.tradetext {
  background: black;
  color: #FDD835;
  height: 100%;
  margin-left: 10px;
  padding: 10px;
  cursor: pointer;
}
</style>
