import { create } from 'zustand';
import type {
  ServerInfo,
  DomainInfo,
  TestParams,
  TestResultsData,
  NetworkAdapterInfo,
  BenchmarkProgress,
} from '../types';
import {
  GetCurrentPreset,
  GetServerList,
  GetDomainList,
  GetTestParams,
  GetLastResults,
  GetNetworkAdapters,
  IsAdmin,
  GetSystemTheme,
} from '../../wailsjs/go/backend/AppService';
import { applyTheme } from '../styles/theme';
import { getErrorMessage } from '../utils/errors';

interface AppState {
  // 预设
  currentPreset: string;
  servers: ServerInfo[];
  domains: DomainInfo[];

  // 测试参数
  testParams: TestParams;

  // 测试状态
  benchmarkRunning: boolean;
  benchmarkProgress: BenchmarkProgress | null;

  // 测试结果
  results: TestResultsData | null;

  // 系统 DNS
  adapters: NetworkAdapterInfo[];
  isAdmin: boolean;

  // 主题
  theme: 'light' | 'dark';

  // 全局错误提示
  errorMsg: string | null;
}

interface AppActions {
  /** 初始化：从后端加载所有数据 */
  initialize: () => Promise<void>;

  setCurrentPreset: (preset: string) => void;
  setServers: (servers: ServerInfo[]) => void;
  setDomains: (domains: DomainInfo[]) => void;
  setTestParams: (params: TestParams) => void;
  setBenchmarkRunning: (running: boolean) => void;
  setBenchmarkProgress: (progress: BenchmarkProgress | null) => void;
  setResults: (results: TestResultsData | null) => void;
  setAdapters: (adapters: NetworkAdapterInfo[]) => void;
  setIsAdmin: (isAdmin: boolean) => void;
  setTheme: (theme: 'light' | 'dark') => void;
  setError: (msg: string) => void;
  clearError: () => void;
}

export const useAppStore = create<AppState & AppActions>()((set) => ({
  // --- 初始状态 ---
  currentPreset: 'cn',
  servers: [],
  domains: [],
  testParams: {
    queries: 10,
    warmup: 1,
    concurrency: 20,
    timeout: 2.0,
  },
  benchmarkRunning: false,
  benchmarkProgress: null,
  results: null,
  adapters: [],
  isAdmin: false,
  theme: 'light',
  errorMsg: null,

  // --- Actions ---

  initialize: async () => {
    try {
      const [
        preset,
        servers,
        domains,
        testParams,
        lastResults,
        adapters,
        isAdmin,
        systemTheme,
      ] = await Promise.all([
        GetCurrentPreset(),
        GetServerList(),
        GetDomainList(),
        GetTestParams(),
        GetLastResults(),
        GetNetworkAdapters().catch(() => [] as NetworkAdapterInfo[]),
        IsAdmin().catch(() => false),
        GetSystemTheme().catch(() => 'light'),
      ]);

      const theme = (systemTheme === 'dark' ? 'dark' : 'light') as 'light' | 'dark';
      applyTheme(theme);

      set({
        currentPreset: preset,
        servers,
        domains,
        testParams,
        results: lastResults ?? null,
        adapters,
        isAdmin,
        theme,
      });
    } catch (err) {
      set({ errorMsg: getErrorMessage(err, '应用初始化失败，请重试。') });
    }
  },

  setCurrentPreset: (preset) => set({ currentPreset: preset }),
  setServers: (servers) => set({ servers }),
  setDomains: (domains) => set({ domains }),
  setTestParams: (params) => set({ testParams: params }),
  setBenchmarkRunning: (running) => set({ benchmarkRunning: running }),
  setBenchmarkProgress: (progress) => set({ benchmarkProgress: progress }),
  setResults: (results) => set({ results }),
  setAdapters: (adapters) => set({ adapters }),
  setIsAdmin: (isAdmin) => set({ isAdmin }),
  setTheme: (theme) => set({ theme }),
  setError: (msg) => set({ errorMsg: msg }),
  clearError: () => set({ errorMsg: null }),
}));
