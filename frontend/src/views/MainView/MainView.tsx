import styles from './MainView.module.css';
import type { ServerData } from '../../features/servers/ServerList/ServerList';
import type { StatsData } from '../../features/stats/StatsBar/StatsBar';
import type { ProxyInfoData } from '../../features/settings/ProxyInfo/ProxyInfo';

import TopBar from '../../features/settings/TopBar';
import StatusArea from '../../features/connection/StatusArea';
import ServerCard from '../../features/servers/ServerCard';

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
