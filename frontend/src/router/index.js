import Vue from 'vue'
import Router from 'vue-router'
import Dash from '@/components/Dash'
import Trading from '@/components/Trading'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Dash',
      component: Dash
    },
    {
      path: '/trackers',
      name: 'Trading',
      component: Trading
    }
  ]
})
