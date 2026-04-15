import { useEffect, useState } from 'preact/hooks';
import styles from './App.module.css';
import { useConnectionStore } from './stores/useConnectionStore';
import { useServerStore } from './stores/useServerStore';
import { useStatsStore } from './stores/useStatsStore';

import MainView from './views/MainView/MainView';
import ServersView from './views/ServersView/ServersView';
import InstallView from './views/InstallView/InstallView';

export type View = 'main' | 'servers';

function App() {
  const [view, setView] = useState<View>('main');
  
  const conn = useConnectionStore();
  const serv = useServerStore();
  const stats = useStatsStore();

  useEffect(() => {
    conn.init();
    serv.init();
    serv.loadServers();
    stats.startPolling(() => useConnectionStore.getState().connected);
  }, []);

  const handleServerSelect = async (i: number) => {
    if (conn.connecting || conn.isDisconnecting || serv.pinging.size > 0) return;
    
    if (conn.connected) {
      await conn.switchServer(i);
    } else {
      await serv.selectServer(i);
    }
    setView('main');
  };

  if (conn.isInstallingCore) {
    return (
      <div className={styles.app}>
        <InstallView 
          statusText={conn.statusText} 
          progress={conn.installProgress} 
        />
      </div>
    );
  }

  return (
    <div className={styles.app}>
      {view === 'main' ? (
        <MainView 
          connected={conn.connected}
          connecting={conn.connecting}
          isDisconnecting={conn.isDisconnecting}
          statusText={conn.statusText}
          useTun={conn.useTun}
          setUseTun={conn.setUseTun}
          systemProxy={conn.systemProxy}
          setSystemProxy={conn.setSystemProxy}
          proxyInfo={conn.proxyInfo}
          toggleConnection={conn.toggleConnection}
          activeServer={serv.servers[serv.selectedServer]}
          stats={stats.stats}
          onOpenServers={() => setView('servers')}
        />
      ) : (
        <ServersView 
          servers={serv.servers}
          selectedServer={serv.selectedServer}
          loading={serv.loading}
          error={serv.error}
          errorMessage={serv.errorMessage}
          pinging={serv.pinging}
          onSelect={handleServerSelect}
          onPing={serv.pingServer}
          onPingAll={serv.pingAll}
          onRefresh={serv.loadServers}
          onBack={() => setView('main')}
        />
      )}
    </div>
  );
}

export default App;
