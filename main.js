'use strict';
const ccxt = require('ccxt');

(async function () {
  var poloniex = new ccxt.poloniex();
  var pmarkets = await poloniex.loadMarkets();
  var bittrex = new ccxt.bittrex();
  var bmarkets = await bittrex.loadMarkets();
  var diffs = []
  for (var currency in pmarkets) {
    if (pmarkets.hasOwnProperty(currency) && bmarkets.hasOwnProperty(currency)) {
      var bprices = await bittrex.fetchOrderBook(currency);
      var pprices = await poloniex.fetchOrderBook(currency);
      if (!bprices.bids[0]) {
        continue;
      }
      var price = bprices.bids[0][0];
      var exchange = 'Bittrex ➼ Poloniex';
      if (pprices.bids[0][0]-bprices.asks[0][0] > bprices.bids[0][0]-pprices.asks[0][0]) {
        price = pprices.bids[0][0];
        exchange = 'Poloniex ➼ Bittrex';
      }
      var diff = Math.max(pprices.bids[0][0]-bprices.asks[0][0],bprices.bids[0][0]-pprices.asks[0][0]);
      if (diff > 0) {
        diffs.push({symbol: currency, diff: diff/price, display: currency + ' '+exchange+': ' + (diff/price*100).toFixed(2).toString()+'%'});
      }
    }
  }
  diffs = diffs.sort((a, b) => {return b.diff-a.diff});
  console.log(diffs)
  for (var i=0; i<diffs.length;i++) {
    console.log(diffs[i].display);
  }
})();
