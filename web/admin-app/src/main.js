import { createApp } from 'vue';
import { createPinia } from 'pinia';
import ElementPlus from 'element-plus';
import zhCn from 'element-plus/es/locale/lang/zh-cn';
import 'element-plus/dist/index.css';
import '@toast-ui/editor/dist/toastui-editor.css';
import './styles.css';
import App from './App.vue';
import router from './router';
import i18n from './i18n';

createApp(App).use(createPinia()).use(router).use(i18n).use(ElementPlus, { locale: zhCn }).mount('#app');
