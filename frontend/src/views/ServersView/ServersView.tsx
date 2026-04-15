import { useState, useMemo, useEffect } from 'preact/hooks';
import styles from './ServersView.module.css';
import { useTranslation } from '../../i18n';
import type { ServerData } from '../../features/servers/ServerList/ServerList';
import Flag from 'react-flagpack';
import Loader from '../../ui/Loader';
import Card from '../../ui/Card';
import Badge from '../../ui/Badge';

interface ServersViewProps {
  servers: ServerData[];
  selectedServer: number;
  loading: boolean;
  error: boolean;
  errorMessage: string;
  pinging: Set<number>;
  onSelect: (index: number) => void;
  onPing: (index: number) => void;
  onPingAll: () => void;
  onRefresh: () => void;
  onBack: () => void;
}

const emojiToCode: Record<string, string> = {
  "🇺🇸": "US", "🇬🇧": "GB", "🇩🇪": "DE", "🇫🇷": "FR", "🇳🇱": "NL",
  "🇷🇺": "RU", "🇺🇦": "UA", "🇰🇿": "KZ", "🇯🇵": "JP", "🇰🇷": "KR",
  "🇸🇬": "SG", "🇭🇰": "HK", "🇹🇼": "TW", "🇦🇺": "AU", "🇨🇦": "CA",
  "🇧🇷": "BR", "🇮🇳": "IN", "🇹🇷": "TR", "🇦🇪": "AE", "🇫🇮": "FI"
};

export default function ServersView({
  servers,
  selectedServer,
  loading,
  pinging,
  onSelect,
  onPingAll,
  onRefresh,
  onBack
}: ServersViewProps) {
  const { t, getCountryName } = useTranslation();
  const [search, setSearch] = useState('');
  const [favorites, setFavorites] = useState<Set<string>>(new Set());

  useEffect(() => {
    const saved = localStorage.getItem('favorites');
    if (saved) {
      try { setFavorites(new Set(JSON.parse(saved))); } catch {}
    }
  }, []);

  const toggleFavorite = (e: any, link: string) => {
    e.stopPropagation();
    const next = new Set(favorites);
    if (next.has(link)) next.delete(link); else next.add(link);
    setFavorites(next);
    localStorage.setItem('favorites', JSON.stringify(Array.from(next)));
  };

  const filteredServers = useMemo(() => {
    return servers.filter(s => 
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.transport?.toLowerCase().includes(search.toLowerCase())
    );
  }, [servers, search]);

  const [favoriteList, otherList] = useMemo(() => {
    const favs: {s: ServerData, i: number}[] = [];
    const others: {s: ServerData, i: number}[] = [];
    filteredServers.forEach(s => {
      const item = { s, i: servers.indexOf(s) };
      if (favorites.has(s.link || '')) favs.push(item); else others.push(item);
    });
    return [favs, others];
  }, [filteredServers, favorites, servers]);

  const renderServerItem = ({ s, i }: {s: ServerData, i: number}) => {
    const code = emojiToCode[s.flag];
    const isSelected = i === selectedServer;
    const isFavorite = favorites.has(s.link || '');

    return (
      <Card 
        key={s.link} 
        className={`${styles.serverItem} ${isSelected ? styles.selected : ''}`}
        onClick={() => onSelect(i)}
      >
        <div className={styles.flagBox}>
          {code ? <Flag code={code} size="l" /> : <span className={styles.emoji}>{s.flag}</span>}
        </div>
        <div className={styles.info}>
          <span className={styles.name}>{getCountryName(s.flag, s.name)}</span>
          <Badge text={s.transport?.toUpperCase() || 'VLESS'} />
        </div>
        <div className={styles.actionArea}>
          <span className={`${styles.latency} ${s.latency_ms === -1 ? styles.error : ''}`}>
            {pinging.has(i) ? <Loader size="small" color="green" /> : 
             s.latency_ms && s.latency_ms > 0 ? `${s.latency_ms} ms` : 
             s.latency_ms === -1 ? 'Error' : '--'}
          </span>
          <button 
            className={`${styles.favBtn} ${isFavorite ? styles.isFav : ''}`}
            onClick={(e) => toggleFavorite(e, s.link || '')}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill={isFavorite ? "currentColor" : "none"} stroke="currentColor" strokeWidth="2.5">
              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
            </svg>
          </button>
        </div>
      </Card>
    );
  };

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <button className={styles.backBtn} onClick={onBack}>
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
            <path d="M15 18l-6-6 6-6" />
          </svg>
        </button>
        <span className={styles.title}>{t('app.title')}</span>
        <button className={styles.refreshBtn} onClick={onRefresh} disabled={loading}>
          {loading ? <Loader size="small" /> : (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
              <path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
            </svg>
          )}
        </button>
      </div>

      <div className={styles.searchBar}>
        <div className={styles.searchInner}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
          </svg>
          <input 
            type="text" 
            placeholder={t('app.search')} 
            value={search}
            onInput={(e) => setSearch(e.currentTarget.value)}
          />
        </div>
      </div>

      <div className={styles.scrollArea}>
        {favoriteList.length > 0 && (
          <div className={styles.section}>
            <span className={styles.sectionTitle}>{t('app.favorites')}</span>
            {favoriteList.map(renderServerItem)}
          </div>
        )}

        <div className={styles.section}>
          <div className={styles.sectionHeader}>
            <span className={styles.sectionTitle}>{t('app.all_servers')}</span>
            <button className={styles.pingAllBtn} onClick={onPingAll} title={t('app.ping_all')}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
                <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
              </svg>
            </button>
          </div>
          {otherList.map(renderServerItem)}
        </div>

        {loading && servers.length === 0 && <div className={styles.empty}><Loader size="large" /></div>}
        {!loading && filteredServers.length === 0 && <div className={styles.empty}><span>{t('app.no_servers')}</span></div>}
      </div>
    </div>
  );
}
