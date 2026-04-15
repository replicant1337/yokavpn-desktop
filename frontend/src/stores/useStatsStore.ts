import { create } from 'zustand';
import * as api from '../../wailsjs/go/main/App';
import type { StatsData } from '../features/stats/StatsBar/StatsBar';

interface StatsState {
  stats: StatsData;
  startPolling: (isConnected: () => boolean) => void;
}

export const useStatsStore = create<StatsState>((set) => ({
  stats: { upload: 0, download: 0 },
  startPolling: (isConnected) => {
    setInterval(async () => {
      if (isConnected()) {
        const stats = await api.GetStats();
        if (stats) {
          set({ stats: { upload: stats.upload_bytes, download: stats.download_bytes } });
        }
      }
    }, 1000);
  }
}));
