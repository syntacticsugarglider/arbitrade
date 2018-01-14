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
        trade is from <strong>{{trackers[symbol].e1}}</strong> for <strong>{{trackers[symbol].ob1.asks[0][0]}}{{trackers[symbol].c2}}</strong>
      </div>
      to
      <div class="path">
        <strong>{{trackers[symbol].e2}}</strong> at <strong>{{trackers[symbol].ob2.bids[0][0]}}{{trackers[symbol].c2}}</strong>
      </div>
      for a profit of
      <div class="path">
        <div class="profit">{{(((trackers[symbol].ob2.bids[0][0]-trackers[symbol].ob1.asks[0][0])/trackers[symbol].ob1.asks[0][0])*100).toFixed(1)}}%</div>
      </div>
      <div class="tradebutton">
        <div @input="(val) => amt(val,symbol)" contenteditable="true" class="tradeinput">0</div>{{trackers[symbol].c2}}<div class="tradetext">TRADE</div>
      </div>
      <div class="book">
        <div class="bookleft">
          <div class="entry ex">{{trackers[symbol].e1}}</div>
          <div :class="actives[symbol] && actives[symbol].indexOf(index)!=-1 ? 'active entry' : 'entry'" v-for="(order, index) in trackers[symbol].ob1.asks"><div>{{order[0]}}</div><div class="small">{{order[1]}}{{trackers[symbol].c1}}</div></div>
        </div>
        <div class="bookright">
          <div class="entry ex">{{trackers[symbol].e2}}</div>
          <div :class="sactives[symbol] && sactives[symbol].indexOf(index)!=-1 ? 'active entry' : 'entry'" v-for="(order, index) in trackers[symbol].ob2.bids"><div>{{order[0]}}</div><div class="small">{{order[1]}}{{trackers[symbol].c1}}</div></div>
        </div>
      </div>
      <div class="rev">Estimated revenue: {{rev[symbol] ? rev[symbol].toFixed(5) : 0}}{{trackers[symbol].c2}}</div>
      <div class="p2">Real profit: {{((rev[symbol]/val[symbol])*100).toFixed(2) == 'NaN' ? 0 : ((rev[symbol]/val[symbol])*100).toFixed(2)}}%</div>
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
      currentTime: Date.now(),
      actives: {},
      sactives: {},
      amount: {},
      rev: {},
      val: {}
    }
  },
  mounted() {
    setInterval(() => {
      this.currentTime = Date.now()
    },1)
  },
  methods: {
    amt: function(val, symbol) {
      if (!val) {
        return;
      }
      if (val.target) {
        var value = parseFloat(val.target.innerHTML);
        this.val[symbol] = value;
      }
      else {
        var value = val;
      }
      var buy = this.$props.trackers[symbol].ob1.asks
      this.actives[symbol] = [];
      var total = 0;
      var tokens = 0;
      for (var i = 0; i<buy.length; i++) {
        if (total<value) {
          var temptotal = total + buy[i][0]*buy[i][1];
          if (temptotal>=value) {
            tokens += (value-total)/buy[i][0];
          }
          else {
            tokens += buy[i][1];
          }
          total = temptotal;
          this.actives[symbol].push(i);
        }
        else {
          break;
        }
      }
      var sell = this.$props.trackers[symbol].ob2.bids
      this.sactives[symbol] = [];
      var stotal = 0;
      var returns = 0;
      for (var i = 0; i<sell.length; i++) {
        if (stotal<tokens) {
          console.log(stotal, tokens)
          if ((stotal+sell[i][1])>=tokens) {
            console.log(tokens-stotal)
            returns += (tokens-stotal)*sell[i][0];
          }
          else {
            returns += sell[i][1]*sell[i][0];
          }
          stotal += sell[i][1];
          this.sactives[symbol].push(i);
        }
        else {
          break;
        }
      }
      console.log(returns, value);
      this.rev[symbol] = returns-value;
    }
  },
  created() {
    this.$parent.$on('update', (data) => {
      this.hTracked.map(symbol => {
        this.amt(this.val[symbol], symbol);
      });
    });
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
  font-size: 2rem;
  margin-bottom: 10px;
  font-family: Futura;
}
.profit2 {
  font-size: 2rem;
  margin-top: 10px;
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
.entry:not(.ex) {
  cursor: pointer;
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
.book {
  width: 100%;
  display: flex;
  justify-content: flex-start;
  flex-flow: column nowrap;
  align-items: flex-start;
  margin-top: 20px;
  overflow-y: scroll;
  height: 100px;
  flex-flow: row nowrap;
}
.bookleft {
  width: 50%;
  display: flex;
  justify-content: flex-start;
  align-items: flex-end;
  flex-flow: column nowrap;
  min-height: 20px;
  border: 2px solid black;
  border-radius: 5px;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  padding-right: 10px;
  border-right: none;
  padding-bottom: 10px;
}
.bookright {
  width: 50%;
  display: flex;
  justify-content: flex-start;
  align-items: flex-start;
  flex-flow: column nowrap;
  min-height: 20px;
  padding-left: 10px;
  border: 2px solid black;
  border-radius: 5px;
  border-top-left-radius: 0;
  padding-bottom: 10px;
  border-bottom-left-radius: 0;
}
.entry {
  font-family: Futura;
}
.entry.ex {
  border-bottom: 2px solid black;
  margin-bottom: 10px;
}
.small {
  font-size: 0.5rem;
  border-bottom: 1px solid black;
}
.active.entry {
  background: black;
  color: white;
}
.rev {
  font-family: Futura;
  margin-top: 15px;
}
.p2 {
  font-family: Futura;
}
</style>
