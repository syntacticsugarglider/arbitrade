'use strict';
const ccxt = require('ccxt');
const to = require('./to.js');
const app = require('express')();
const http = require('http').Server(app);
const io = require('socket.io')(http);
const { performance } = require('perf_hooks');
var fetch = require('node-fetch');

io.on('connection', function(socket){
  console.log('Web interface connected');
});

http.listen(3001, 'localhost', function(){
  console.log('Realtime update socket listening on *:3001');
});

async function connectExchanges(ts) {
  var targets = ts;
  var exchanges = {};
  var promises = targets.map((target) => {
    return new Promise(async function(resolve, reject) {
      exchanges[target] = new ccxt[target]();
      exchanges[target].timeout = 2000;
      return resolve();
    });
  });
  await Promise.all(promises);
  return exchanges;
}

async function getSymbols (xcs) {
  var exchanges = xcs
  var symbols = []
  var errored=[]
  for (var i=0; i<exchanges.length; i++) {
    let err, foo
    [err, foo] = await to(exchanges[i].loadMarkets());
    if (err) {
      console.log(`Connection to ${exchanges[i].name} failed`);
      errored.push(exchanges[i]);
      continue;
    }
    var _symbols = exchanges[i].symbols;
    for (var j=0; j<_symbols.length; j++) {
      if (symbols.indexOf(_symbols[j]) === -1) {
        symbols.push(_symbols[j]);
      }
    }
  }
  for (var i=0; i<errored.length; i++) {
    exchanges.splice(exchanges.indexOf(errored[i]),1);
  }
  errored = []
  for (var i=0; i<symbols.length; i++) {
    symbols[i] = {
      symbol: symbols[i],
      exchanges: {},
      c1: symbols[i].split('/')[0],
      c2: symbols[i].split('/')[1],
      diffs: []
    };
    for (var j=0; j<exchanges.length; j++) {
      if (exchanges[j].symbols.indexOf(symbols[i].symbol)!=-1) {
        symbols[i].exchanges[exchanges[j].id]={};
      }
    }
    if (Object.keys(symbols[i].exchanges).length < 2) {
      errored.push(symbols[i]);
    }
  }
  for (var i=0; i<errored.length; i++) {
    symbols.splice(symbols.indexOf(errored[i]),1);
  }
  return symbols;
}

/*async function pGetPrices(x, s) {
  var exchanges = x;
  var symbols = s;
  var err;
  var v;
  var promises = symbols.map((symbol) => {
    return new Promise(async function(resolve, reject) {
      var ipromises = Object.keys(symbol.exchanges).map((exchange) => {
        return new Promise(async function(resolve, reject) {
          if (exchanges.hasOwnProperty(exchange)) {
            if (!exchanges[exchange].tickers || Object.keys(exchanges[exchange].tickers).length===0) {
              var tickers;
              [err, tickers] = await to(exchanges[exchange].fetchTickers());
              exchanges[exchange].tickers = tickers;
              if (err) {
                delete symbol.exchanges[exchange];
                return resolve([undefined, undefined]);
              }
              return resolve([symbols.indexOf(symbol), exchange]);
            }
            else {
              return resolve([symbols.indexOf(symbol), exchange]);
            }
          }
          else {
            return resolve([undefined, undefined]);
          }
          return resolve([undefined, undefined]);
        }).then((s) => {
          return new Promise((resolve, reject) => {
            try {
              var e = s[1];
              s = s[0];
              if (s===undefined) {
                return resolve([undefined,undefined]);
              }
              symbols[s].exchanges[e].timestamp = exchanges[e].tickers[symbols[s].symbol].timestamp;
              symbols[s].exchanges[e].bid = exchanges[e].tickers[symbols[s].symbol].bid;
              symbols[s].exchanges[e].ask = exchanges[e].tickers[symbols[s].symbol].ask;
              return resolve([s, e]);
            }
            catch (e) {
              console.log(e);
              return resolve([undefined, undefined]);
            }
            return resolve([undefined, undefined]);
          });
        });
      });
      await Promise.all(ipromises);
      return resolve();
    });
  });
  await Promise.all(promises);
  for (var i=0; i<exchanges.length; i++) {
    delete exchanges[i].tickers;
  }
  return symbols;
}*/

async function getPrices(x) {
  var exchanges = x;
  var symbols = [];
  var err;
  [err, symbols] = await to(getSymbols(Object.values(exchanges)));
  if (err) {
    console.log(err);
    return;
  }
  for (var symbol in symbols) {
    if (symbols.hasOwnProperty(symbol)) {
      for (var exchange in symbols[symbol].exchanges) {
        if (symbols[symbol].exchanges.hasOwnProperty(exchange)) {
          var ticker;
          let err;
          if (exchanges[exchange].tickers) {
            ticker = exchanges[exchange].tickers[symbols[symbol].symbol];
            if (!ticker) {
              [err, ticker] = await to(exchanges[exchange].fetchTicker(symbols[symbol].symbol));
              if (err) {
                delete symbols[symbol].exchanges[exchange];
                console.log(`Single ticker access failed to ${exchanges[exchange].name} on market ${symbols[symbol].symbol}`);
                continue;
              }
            }
          }
          else {
            [err, ticker] = await to(exchanges[exchange].fetchTicker(symbols[symbol].symbol));
            if (err) {
              delete symbols[symbol].exchanges[exchange];
              console.log(`Single ticker access failed to ${exchanges[exchange].name} on market ${symbols[symbol].symbol}`);
              continue;
            }
          }
          symbols[symbol].exchanges[exchange].bid = ticker.bid;
          symbols[symbol].exchanges[exchange].ask = ticker.ask;
          symbols[symbol].exchanges[exchange].timestamp = ticker.timestamp;
          if ((ticker.ask === 0) || (ticker.bid === 0)) {
            delete symbols[symbol].exchanges[exchange];
            console.log(`0-valued bid or ask from ${exchanges[exchange].name} on market ${symbols[symbol].symbol}`);
          }
        }
      }
    }
  }
  for (var symbol in symbols) {
    if (symbols.hasOwnProperty(symbol)) {
      for (var exchange in symbols[symbol].exchanges) {
        if (symbols[symbol].exchanges.hasOwnProperty(exchange)) {
          for (var xchange in symbols[symbol].exchanges) {
            if (symbols[symbol].exchanges.hasOwnProperty(exchange) && (xchange != exchange)) {
              if (symbols[symbol].exchanges[exchange].ask < symbols[symbol].exchanges[xchange].bid) {
                var tempDiff = {
                  ex1: exchange,
                  ex2: xchange,
                  diff: symbols[symbol].exchanges[xchange].bid - symbols[symbol].exchanges[exchange].ask,
                  rdiff: (symbols[symbol].exchanges[xchange].bid - symbols[symbol].exchanges[exchange].ask)/symbols[symbol].exchanges[xchange].bid
                };
                symbols[symbol].diffs.push(tempDiff);
              }
            }
          }
        }
      }
    }
  }
  for (var symbol in symbols) {
    if (symbols.hasOwnProperty(symbol)) {
      symbols[symbol].diffs.sort((a, b)=>{return b.rdiff-a.rdiff});
    }
  }
  symbols.sort((a, b) => {
    var bdiff = b.diffs.length > 0 ? b.diffs[0].rdiff : undefined
    var adiff = a.diffs.length > 0 ? a.diffs[0].rdiff : undefined
    if (!adiff) {
      if (!bdiff) {
        return 0
      }
      else {
        return 1
      }
    }
    else if (!bdiff) {
      return -1
    }
    return bdiff-adiff
  });
  return symbols
}

(async function () {
  var symbols, exchanges, err;
  [err, exchanges] = await to(connectExchanges([
    'bleutrade', 'hitbtc', 'cryptopia'
  ]));
  [err, symbols] = await to(getSymbols(Object.values(exchanges)));
  symbols = await getPrices(exchanges);
  /*var poloniexErrored = []
  for (var i=0; i<symbols.length; i++) {
    if (symbols[i].diffs.length>0) {
      if (symbols[i].diffs[0].ex1==='poloniex' || symbols[i].diffs[0].ex2==='poloniex') {
        console.log(`Verifying operability of poloniex market ${symbols[i].symbol}`);
        await fetch(`https://poloniex.com/exchange#${symbols[i].diffs[0].c2}_${symbols[i].diffs[0].c1}`).then(function(res) {
          return res.text();
        }).then(function(body) {
          console.log(body)
          if (body.indexOf('currently under maintenance')!=-1) {
            console.log(`Culled poloniex market ${symbols[i].symbol} due to outage`)
          }
        });
      }
    }
  }
  console.log('Poloniex market verification complete')*/
  while (true) {
    symbols = await getPrices(exchanges);
    io.emit('symbols',symbols);
  }
})();
