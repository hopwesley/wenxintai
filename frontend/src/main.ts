import {createApp} from 'vue'
import {createPinia} from 'pinia'
import App from './App.vue'
import {router} from './controller/router_index'
import '@/styles/base.css'

import VChart from 'vue-echarts'
import {use} from 'echarts/core'
import {CanvasRenderer} from 'echarts/renderers'
import {RadarChart, BarChart} from 'echarts/charts'
import {LegendComponent, TooltipComponent, GridComponent} from 'echarts/components'

use([
    CanvasRenderer,
    RadarChart,
    BarChart,
    LegendComponent,
    TooltipComponent,
    GridComponent,
])

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)   // ← 先挂 Pinia
app.use(router)  // ← 再挂 Router（顺序和这个没硬性要求，只要在 mount 前）
app.component('VChart', VChart)
app.mount('#app')
