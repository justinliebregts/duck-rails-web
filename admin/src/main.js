import 'core-js/stable'
import Vue from 'vue'
import App from './App'
import router from './router'
import CoreuiVue from '@coreui/vue'
import { iconsSet as icons } from './assets/icons/icons.js'
import store from './store'
import axios from 'axios'
import VueAxios from 'vue-axios'
import CoreuiVueCharts from '@coreui/vue-chartjs'

import VueCompositionAPI from '@vue/composition-api' // The Bridge
import { createPinia, PiniaVuePlugin } from 'pinia' // for store


// for display json
import JsonViewer from 'vue-json-viewer'

Vue.config.performance = true
Vue.use(CoreuiVue)
Vue.use(VueAxios, axios)
Vue.use(CoreuiVueCharts)

// Import JsonViewer as a Vue.js plugin
Vue.use(JsonViewer)

// Tell Vue 2 to use Pinia
Vue.use(VueCompositionAPI) // Bridge first
Vue.use(PiniaVuePlugin)
const pinia = createPinia()

Vue.prototype.$log = console.log.bind(console)

// THIS IS THE KEY:
import { useAuthStore } from './stores/auth' // Import your specific store
Vue.prototype.$auth = useAuthStore(pinia)

import gstate from './global-state.js'
Vue.prototype.$gstate = gstate;
Vue.prototype.noop = function() {};

document.title = 'Gun Raiders Classic Admin';

new Vue({
    el: '#app',
    router,
    store,
    pinia, // Inject Pinia
    icons,
    template: '<App/>',
    components: {
        App
    }
})