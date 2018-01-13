<template>
  <div class="container">
    <div v-for="opportunity, index in symbols" v-if="opportunity.diffs.length>0" class="opportunity">
      <span class="yellow">#{{index+1}}&nbsp;</span>Buy&nbsp;<span class="yellow">{{opportunity.c1}}</span>&nbsp;with&nbsp;<span class="yellow">{{opportunity.c2}}</span>&nbsp;on<span class="yellow">&nbsp;{{opportunity.diffs[0].ex1}}&nbsp;</span>for&nbsp;<span class="yellow">{{opportunity.exchanges[opportunity.diffs[0].ex1].ask}}</span>,
      and sell on<span class="yellow">&nbsp;{{opportunity.diffs[0].ex2}}</span>&nbsp;for&nbsp;<span class="yellow">{{opportunity.exchanges[opportunity.diffs[0].ex2].bid}}</span>&nbsp; to a profit of&nbsp;<span class="yellow">{{(opportunity.diffs[0].rdiff*100).toFixed(1)}}%</span>&nbsp;
      ({{Math.ceil((Math.floor(currentTime)-Math.max(opportunity.exchanges[opportunity.diffs[0].ex2].timestamp,opportunity.exchanges[opportunity.diffs[0].ex1].timestamp)))}}ms old)
    </div>
  </div>
</template>

<script>
export default {
  name: 'dash',
  props: ['symbols'],
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
  margin-top: 60px;
}
.opportunity {
  height: 40px;
  color: white;
  width: 100%;
  background: rgba(255,255,255,0.1);
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: center;
  font-family: Futura;
  padding-left: 20px;
  padding-top: 10px;
  padding-bottom: 10px;
  padding-right: 20px;
  font-weight: 500;
  border-bottom: 1px solid #FDD835;
}
.yellow {
  color: #FDD835;
  font-weight: 800;
}
</style>
