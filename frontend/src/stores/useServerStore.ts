import { create } from 'zustand';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import * as api from '../../wailsjs/go/main/App';
import type { ServerData } from '../components/ServerList/ServerList';

interface ServerState {
  servers: ServerData[];
  selectedServer: number;
  loading: boolean;
  error: boolean;
  errorMessage: string;
  pinging: Set<number>;

  loadServers: () => Promise<void>;
  selectServer: (index: number) => Promise<void>;
  pingServer: (index: number) => Promise<void>;
  pingAll: () => Promise<void>;
  init: () => void;
}

export const useServerStore = create<ServerState>((set, get) => ({
  servers: [],
  selectedServer: 0,
  loading: true,
  error: false,
  errorMessage: '',
  pinging: new Set(),

  init: () => {
    EventsOn('best-server-selected', (index: number) => {
      set({ selectedServer: index });
    });
  },


  loadServers: async () => {
    set({ loading: true, error: false });
    try {
      const result = await api.FetchSubscription('remnawave', 'https://panel.umbr.rest/api/sub/p-aPhsZHR53TPjtk');
      if (result.success && result.servers) {
        set({ servers: result.servers });
      }
    } catch (e) {
      set({ error: true, errorMessage: 'Failed to load servers' });
    } finally {
      set({ loading: false });
    }
  },

  selectServer: async (index) => {
    set({ selectedServer: index });
    await api.SetActiveServer(index);
  },

  pingServer: async (index) => {
    const { servers, pinging } = get();
    if (!servers[index]) return;

    set({ pinging: new Set(pinging).add(index) });
    try {
      const result = await api.TestServer(servers[index].link, index, servers[index].name);
      if (result) {
        const nextServers = [...servers];
        nextServers[index] = { ...nextServers[index], latency_ms: result.latency_ms };
        set({ servers: nextServers });
      }
    } finally {
      const nextPinging = new Set(get().pinging);
      nextPinging.delete(index);
      set({ pinging: nextPinging });
    }
  },

  pingAll: async () => {
    const { servers, pingServer } = get();
    for (let i = 0; i < servers.length; i++) {
      await pingServer(i);
    }
  }
}));
