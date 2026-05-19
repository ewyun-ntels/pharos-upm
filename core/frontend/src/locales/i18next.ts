import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import en from './en/translation.json';
import ko from './ko/translation.json';

i18n.use(LanguageDetector)
    .use(initReactI18next)
    .init({
      resources: {
        ko: {
          translation: ko,
        },
        en: {
          translation: en,
        },
      },
      supportedLngs: ['ko', 'en'],
      fallbackLng: 'en',
      load: 'languageOnly',
      nonExplicitSupportedLngs: true,
      detection: {
        // 감지 우선순위
        order: ['navigator', 'localStorage', 'htmlTag'],

        // 감지 결과를 저장할 위치 (브라우저에서만 적용)
        caches: ['localStorage'],
      },
      interpolation: {
        escapeValue: false,
      },
    });

i18n.on('languageChanged', (lng) => {
  const short = lng.split('-')[0];
  if (lng !== short && i18n.hasResourceBundle(short, 'translation')) {
    i18n.changeLanguage(short);
  }
});

export default i18n;
