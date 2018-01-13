'use strict';
const ccxt = require('ccxt');
const to = require('./to.js');
const app = require('express')();
const http = require('http').Server(app);
const io = require('socket.io')(http);
const { performance } = require('perf_hooks');
const zombie = require('zombie')
const fetch = require('node-fetch');

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
      exchanges[target].timeout = 5000;
      return resolve();
    });
  });
  await Promise.all(promises);
  return exchanges;
}

async function getSymbols(xcs) {
  var exchanges = xcs
  var err, foo;
  var symbols = []
  var promises = Object.values(exchanges).map((exchange => {
    var promise = new Promise(async (resolve, reject) => {
      [err, foo] = await to(exchange.loadMarkets());
      if (err) {
        return reject(exchange);
      }
      return resolve(exchange.symbols);
    })
    promise.then((data) => {
      return new Promise(async (resolve, reject) => {
        data.map((symbol) => {
          if (symbols.indexOf(symbol) === -1) {
            symbols.push(symbol);
          }
        });
        return resolve();
      });
    }).catch((err) => {
      console.log(`Failed to get market data from ${err.name}`);
    })
    return promise;
  }));
  await Promise.all(promises.map(p => p.catch(() => undefined)));
  symbols = symbols.map((symbol) => {
    var _exchanges = Object.values(exchanges).filter(exchange => exchange.symbols && exchange.symbols.indexOf(symbol) != -1)
    return {
      symbol: symbol,
      c1: symbol.split('/')[0],
      c2: symbol.split('/')[1],
      exchanges: _exchanges
    };
  });
  symbols = symbols.filter(symbol => Object.keys(symbol.exchanges).length > 1);
  return symbols;
}

(async function () {
  var symbols, exchanges, err;
  exchanges = await connectExchanges(['bleutrade', 'hitbtc', 'cryptopia']);
  symbols = await getSymbols(exchanges);
  console.log(symbols.length)
})();
