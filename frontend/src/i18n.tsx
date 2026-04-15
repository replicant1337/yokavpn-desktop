import { createContext } from 'preact';
import { useContext, useState, useCallback, useMemo } from 'preact/hooks';
import ru from './locales/ru.json';
import en from './locales/en.json';

const translations: Record<string, any> = { ru, en };

const emojiToKeyMap: Record<string, string> = {
  "🇷🇺": "countries.russia",
  "🇩🇪": "countries.germany",
  "🇳🇱": "countries.netherlands",
  "🇺🇸": "countries.usa",
  "🇬🇧": "countries.uk",
  "🇫🇷": "countries.france",
  "🇹🇷": "countries.turkey",
  "🇫🇮": "countries.finland",
  "🇰🇿": "countries.kazakhstan",
  "🇸🇬": "countries.singapore",
  "🇭🇰": "countries.hong kong",
  "🇯🇵": "countries.japan",
  "🇦🇪": "countries.uae",
  "🇺🇦": "countries.ukraine",
  "🇰🇷": "countries.south korea",
  "🇹🇼": "countries.taiwan",
  "🇦🇺": "countries.australia",
  "🇨🇦": "countries.canada",
  "🇧🇷": "countries.brazil",
  "🇮🇳": "countries.india",
  "🇵🇱": "countries.poland",
  "🇦🇹": "countries.austria",
  "🇸🇪": "countries.sweden"
};

interface TranslationContextType {
  t: (path: string, params?: Record<string, any>) => string;
  lang: string;
  changeLanguage: (newLang: string) => void;
  availableLanguages: string[];
  getCountryName: (emoji: string, defaultName: string) => string;
}

const TranslationContext = createContext<TranslationContextType | undefined>(undefined);

export function TranslationProvider({ children }: { children: any }) {
  const [lang, setLang] = useState(localStorage.getItem('lang') || 'ru');

  const t = useCallback((path: string, params?: Record<string, any>) => {
    const keys = path.split('.');
    let result = translations[lang];
    
    for (const key of keys) {
      if (!result || result[key] === undefined) return path;
      result = result[key];
    }
    
    let text = result as string;
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        text = text.replace(`{${key}}`, String(value));
      });
    }
    
    return text;
  }, [lang]);

  const changeLanguage = (newLang: string) => {
    if (translations[newLang]) {
      setLang(newLang);
      localStorage.setItem('lang', newLang);
    }
  };

  const getCountryName = useCallback((emoji: string, defaultName: string) => {
    const key = emojiToKeyMap[emoji];
    if (key) {
      const translated = t(key);
      if (translated !== key) return translated;
    }
    return defaultName;
  }, [t]);

  const availableLanguages = useMemo(() => Object.keys(translations), []);

  const value = { t, lang, changeLanguage, availableLanguages, getCountryName };

  return (
    <TranslationContext.Provider value={value}>
      {children}
    </TranslationContext.Provider>
  );
}

export function useTranslation() {
  const context = useContext(TranslationContext);
  if (context === undefined) {
    throw new Error('useTranslation must be used within a TranslationProvider');
  }
  return context;
}
