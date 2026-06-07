import { createI18n } from 'vue-i18n';
import zhCN from './zh-CN';
import enUS from './en-US';

const messages = {
  'zh-CN': zhCN,
  'en-US': enUS,
};

export const i18n = createI18n({
  legacy: false,
  locale: 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages,
  missingWarn: false,
  fallbackWarn: false,
});

export function setLocale(nextLocale) {
  if (messages[nextLocale]) i18n.global.locale.value = nextLocale;
}

export function t(path, params = {}) {
  return i18n.global.t(path, params);
}

export default i18n;
