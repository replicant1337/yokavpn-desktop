import styles from './MainView.module.css';
import type { ServerData } from '../../components/ServerList/ServerList';
import type { StatsData } from '../../components/StatsBar/StatsBar';
import type { ProxyInfoData } from '../../components/ProxyInfo/ProxyInfo';

import TopBar from './components/TopBar';
import StatusArea from './components/StatusArea';
import ServerCard from './components/ServerCard';

interface MainViewProps {
  connected: boolean;
  connecting: boolean;
  isDisconnecting: boolean;
  statusText: string;
  useTun: boolean;
  setUseTun: (val: boolean) => void;
  systemProxy: boolean;
  setSystemProxy: (val: boolean) => void;
  proxyInfo: ProxyInfoData;
  toggleConnection: () => void;
  activeServer?: ServerData;
  stats: StatsData;
  onOpenServers: () => void;
}

export default function MainView({
  connected,
  connecting,
  isDisconnecting,
  statusText,
  useTun,
  setUseTun,
  systemProxy,
  setSystemProxy,
  proxyInfo,
  toggleConnection,
  activeServer,
  stats,
  onOpenServers
}: MainViewProps) {
  return (
    <div className={styles.container}>
      <TopBar 
        useTun={useTun} 
        setUseTun={setUseTun} 
        disabled={connected || connecting || isDisconnecting} 
      />

      <StatusArea 
        connected={connected}
        connecting={connecting}
        isDisconnecting={isDisconnecting}
        statusText={statusText}
        stats={stats}
        toggleConnection={toggleConnection}
      />

      <ServerCard 
        activeServer={activeServer}
        connected={connected}
        onOpenServers={onOpenServers}
      />
    </div>
  );
}
