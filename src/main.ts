import { createApp } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import router from './router'
import './style.css'
import App from './App.vue'
import { registerCorePlugins } from '@/core/plugins/pluginLoader'
import { kpFormsPlugin } from '@/plugins/kp-forms/plugin'
import pdfMake from 'pdfmake/build/pdfmake';
import pdfFonts from 'pdfmake/build/vfs_fonts'

(pdfMake as any).addVirtualFileSystem(pdfFonts);
window.pdfMake = pdfMake

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
setActivePinia(pinia)

registerCorePlugins([
	{ id: 'kpForms', factory: kpFormsPlugin }
])
app.use(router)

app.mount('#app')
