import { create } from 'zustand';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import * as api from '../../wailsjs/go/main/App';
import type { ProxyInfoData } from '../components/ProxyInfo/ProxyInfo';

interface ConnectionState {
  connected: boolean;
  connecting: boolean;
  isDisconnecting: boolean;
  isInstallingCore: boolean;
  installProgress: { percentage: number; label: string };
  statusText: string;
  useTun: boolean;
  systemProxy: boolean;
  proxyInfo: ProxyInfoData;
  retryCount: number;
  maxRetries: number;

  setUseTun: (val: boolean) => void;
  setSystemProxy: (val: boolean) => void;
  setStatusText: (text: string) => void;
  
  connect: () => Promise<void>;
  disconnect: () => Promise<void>;
  toggleConnection: () => Promise<void>;
  switchServer: (index: number) => Promise<void>;
  init: () => void;
}

export const useConnectionStore = create<ConnectionState>((set, get) => ({
  connected: false,
  connecting: false,
  isDisconnecting: false,
  isInstallingCore: false,
  installProgress: { percentage: 0, label: '' },
  statusText: '',
  useTun: false,
  systemProxy: true,
  proxyInfo: { ip: '127.0.0.1', port: 10808 },
  retryCount: 0,
  maxRetries: 3,

  setUseTun: (useTun) => set({ useTun }),
  setSystemProxy: (systemProxy) => set({ systemProxy }),
  setStatusText: (statusText) => set({ statusText }),

  init: () => {
    EventsOn('core-install-state', (isInstalling: boolean) => set({ isInstallingCore: isInstalling }));
    EventsOn('core-install-progress', (progress: any) => set({ installProgress: progress }));
    EventsOn('connect-status', (status: string) => set({ statusText: status }));
    EventsOn('proxy-info-update', (info: any) => set({ proxyInfo: info }));
    
    EventsOn('vpn-state-changed', (state: string) => {
      set({ 
        connecting: state === 'starting',
        isDisconnecting: state === 'disconnecting',
        connected: state === 'connected'
      });
    });

    EventsOn('connection-lost', async () => {
      set({ connected: false });
      const { retryCount, maxRetries, connect } = get();
      if (retryCount < maxRetries) {
        set({ retryCount: retryCount + 1 });
        await new Promise(r => setTimeout(r, 2000));
        await connect();
      }
    });
  },

  connect: async () => {
    const { useTun, systemProxy } = get();
    try {
      await (api as any).Connect(useTun, systemProxy);
    } catch (e) {
      set({ connected: false });
    }
  },

  disconnect: async () => {
    try {
      await api.Disconnect();
    } catch (e) {
      console.error(e);
    }
  },

  toggleConnection: async () => {
    const { connected, connecting, isDisconnecting, connect, disconnect } = get();
    if (connecting || isDisconnecting) return;
    
    if (connected) {
      await disconnect();
    } else {
      set({ retryCount: 0 });
      await connect();
    }
  },

  switchServer: async (index: number) => {
    const { useTun, systemProxy } = get();
    try {
      await (api as any).SwitchServer(index, useTun, systemProxy);
    } catch (e) {
      console.error(e);
    }
  }
}));
