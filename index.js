'use strict';
const ccxt = require('ccxt');
const to = require('./to.js');
const app = require('express')();
const http = require('http').Server(app);
const io = require('socket.io')(http);
const { performance } = require('perf_hooks');
var hTracked = [];
var trackerSemaphores = {};

io.on('connection', function(socket){
  console.log('Web interface connected');
  socket.emit('hTracked', hTracked);
  socket.on('hTrackedP', function(data) {
    hTracked.push(data);
    trackerSemaphores[data] = 'start';
  });
  socket.on('hTrackedM', function(data) {
    hTracked.splice(hTracked.indexOf(data),1);
    trackerSemaphores[data] = 'stop';
  });
});
http.listen(3001, 'localhost', function(){
  console.log('Realtime update socket listening on *:3001');
});

async function log(message) {
  console.log(message);
}

async function connectExchanges(ts) {
  var targets = ts;
  var exchanges = {};
  var promises = targets.map((target) => {
    return new Promise(async function(resolve, reject) {
      exchanges[target] = new ccxt[target]();
      exchanges[target].timeout = 5000;
      exchanges[target].tickers = {};
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
      log(`Failed to get market data from ${err.name}`);
      delete exchanges[err.id];
    })
    return promise;
  }));
  await Promise.all(promises.map(p => p.catch(() => undefined)));
  symbols = symbols.map((symbol) => {
    var _exchanges = Object.values(exchanges).filter(exchange => exchange.symbols && exchange.symbols.indexOf(symbol) != -1);
    var __exchanges = {};
    _exchanges.map((exchange) => {
      __exchanges[exchange.id] = {}
    });
    return {
      symbol: symbol,
      c1: symbol.split('/')[0],
      c2: symbol.split('/')[1],
      exchanges: __exchanges,
      diffs: []
    };
  });
  symbols = symbols.filter(symbol => Object.keys(symbol.exchanges).length > 1);
  return symbols;
}

async function cullOfflineMarkets(s, e) {
  var symbols = s;
  var exchanges = e;
  symbols.map((symbol) => {
    Object.keys(symbol.exchanges).map((exchange) => {
      if (exchanges[exchange].markets[symbol.symbol].info && exchanges[exchange].markets[symbol.symbol].info.StatusMessage && exchanges[exchange].markets[symbol.symbol].info.StatusMessage.length > 1) {
        //console.log(exchanges[exchange].markets[symbol.symbol].info.StatusMessage, symbol.symbol, exchange);
        delete symbol.exchanges[exchange];
      }
      if (exchanges[exchange].markets[symbol.symbol].info && exchanges[exchange].markets[symbol.symbol].info.Status && exchanges[exchange].markets[symbol.symbol].info.Status != 'OK') {
        //console.log(exchanges[exchange].markets[symbol.symbol].info.Status, symbol.symbol, exchange);
        delete symbol.exchanges[exchange];
      }
    });
  });
  symbols = symbols.filter(symbol => Object.keys(symbol.exchanges).length > 1);
  return symbols;
}

async function getPrices(s, e) {
  var symbols = s;
  var exchanges = e;
  var err;
  var promises = Object.keys(exchanges).map((_exchange) => {
    var exchange = exchanges[_exchange];
    return new Promise(async (resolve, reject) => {
      if (exchange.hasFetchTickers) {
        var a, b;
        a = performance.now();
        [err, exchange.tickers] = await to(exchange.fetchTickers());
        b = performance.now();
        log(`Took ${b-a}ms to get prices from ${exchange.name}`)
        if (!err) {
          return resolve();
        }
        return reject();
      }
      return reject();
    });
  });
  await Promise.all(promises.map(p => p.catch(() => undefined)));
  symbols.map((symbol) => {
    Object.keys(symbol.exchanges).map((exchange) => {
      if (exchanges[exchange].tickers && exchanges[exchange].tickers.hasOwnProperty(symbol.symbol)) {
        symbol.exchanges[exchange].bid = exchanges[exchange].tickers[symbol.symbol].bid;
        symbol.exchanges[exchange].ask = exchanges[exchange].tickers[symbol.symbol].ask;
        symbol.exchanges[exchange].timestamp = exchanges[exchange].tickers[symbol.symbol].timestamp;
        if (symbol.exchanges[exchange].bid===0 || symbol.exchanges[exchange].timestamp===0 || symbol.exchanges[exchange].ask===0) {
          delete symbol.exchanges[exchange];
        }
      }
    });
  });
  return symbols;
}

async function getDiffs(s, e) {
  var symbols = s;
  var exchanges = e;
  var diffs = [];
  symbols.map((symbol) => {
    symbol.diffs = [];
    Object.keys(symbol.exchanges).map((_exchange1) => {
      var e1 = symbol.exchanges[_exchange1];
      var _ex1 = exchanges[_exchange1].id;
      var ex1 = exchanges[_exchange1].name;
      Object.keys(symbol.exchanges).map((_exchange2) => {
        var e2 = symbol.exchanges[_exchange2];
        var ex2 = exchanges[_exchange2].name;
        var _ex2 = exchanges[_exchange2].id;
        if (e2 != e1) {
          if (e1.ask < e2.bid) {
            symbol.diffs.push({
              diff: e2.bid-e1.ask,
              rdiff: (e2.bid-e1.ask)/e1.ask,
              timestamp: Math.min(e1.timestamp, e2.timestamp),
              ex1: ex1,
              ex2: ex2,
              id1: _ex1,
              id2: _ex2
            });
          }
        }
      });
    });
    symbol.diffs.sort((a, b) => {
      return b.rdiff - a.rdiff;
    });
  });
  return symbols;
}

async function sortDiffs(s, e) {
  var symbols = s;
  var exchanges = e;
  symbols.sort((a, b) => {
    var _a = (a.diffs.length < 1);
    var _b = (b.diffs.length < 1);
    if (_a && !_b) {
      return 1;
    }
    if (_b && !_a) {
      return -1;
    }
    if (_b && _a) {
      return 0;
    }
    return b.diffs[0].rdiff - a.diffs[0].rdiff;
  });
  return symbols;
}

async function manageHTrackSemaphores(s, e, t) {
  var symbols = s;
  var exchanges = e;
  var trackers = t;
  Object.keys(trackers).map((symbol) => {
    var tracker = trackers[symbol];
    var _symbols = symbols.filter(_symbol => _symbol.symbol===symbol);
    var _symbol = _symbols[0];
    var id1 = _symbol.diffs[0].id1;
    var id2 = _symbol.diffs[0].id2;
    if (trackers[symbol]==='start') {
      trackers[symbol] = 'run';
      new Promise(async (resolve, reject) => {
        while (trackers[symbol] === 'run') {
          var promises = [];
          var err, data1, data2;
          promises.push((async function() {
            [err, data1] = await to(exchanges[id1].fetchOrderBook(symbol));
          })());
          promises.push((async function() {
            [err, data2] = await to(exchanges[id2].fetchOrderBook(symbol));
          })());
          await Promise.all(promises.map(p => p.catch(() => undefined)));
        var tempobj = {
            c1: symbol.split('/')[0],
            c2: symbol.split('/')[1],
            symbol: symbol
          };
          if (data1) {
            tempobj.e1 = id1;
            tempobj.ob1 = data1;
            if (data2) {
              tempobj.timestamp = Math.min(data1.timestamp, data2.timestamp);
            }
          }
          if (data2) {
            tempobj.e2 = id2;
            tempobj.ob2 = data2;
          }
          io.emit('hTrackData', tempobj);
        }
        return resolve();
      });
    }
  });
}

(async function () {
  var symbols, exchanges, err;
  var a, b;
  a = performance.now();
  //exchanges = await connectExchanges(ccxt.exchanges);
  exchanges = await connectExchanges(['liqui', 'cryptopia', 'bleutrade', 'poloniex', 'kraken', 'gdax']);
  b = performance.now();
  log(`Took ${b-a}ms to connect exchanges`);
  a = performance.now();
  symbols = await getSymbols(exchanges);
  b = performance.now();
  log(`Took ${b-a}ms to get symbols`);
  a = performance.now();
  symbols = await cullOfflineMarkets(symbols, exchanges);
  b = performance.now();
  log(`Took ${b-a}ms to cull offline markets`);
  while (true) {
    a = performance.now();
    symbols = await getPrices(symbols, exchanges);
    symbols = await getDiffs(symbols, exchanges);
    symbols = await sortDiffs(symbols, exchanges);
    io.emit('symbols', symbols);
    await to(manageHTrackSemaphores(symbols, exchanges, trackerSemaphores));
    b = performance.now();
    log(`Took ${b-a}ms to complete one cycle of pricing`);
  }
})();
