import { useTranslation as useI18nextTranslation } from 'react-i18next';

// ============================================================
// useTranslation – refine-compatible wrapper around react-i18next
//
// Refine's useTranslation returns { translate, changeLocale, getLocale }
// react-i18next returns { t, i18n }
// This wrapper unifies the API so call-sites work without change.
// ============================================================

export function useTranslation() {
  const { t, i18n } = useI18nextTranslation();

  const translate = (key: string, params?: Record<string, any>): string => {
    const result = t(key, params as any);
    return String(result);
  };

  const changeLocale = (lang: string): Promise<any> => {
    return i18n.changeLanguage(lang);
  };

  const getLocale = (): string => {
    return i18n.language;
  };

  return { translate, changeLocale, getLocale };
}
