import { useEffect } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { useAppStore } from '../store/useAppStore';
import type { BenchmarkProgress, TestResultsData } from '../types';

/**
 * 监听 Wails 后端事件并更新 Zustand Store 的 Hook。
 * 应在根组件 App 中调用一次。
 */
export function useWailsEvents() {
  const setBenchmarkProgress = useAppStore((s) => s.setBenchmarkProgress);
  const setBenchmarkRunning = useAppStore((s) => s.setBenchmarkRunning);
  const setResults = useAppStore((s) => s.setResults);

  useEffect(() => {
    EventsOn('benchmark:progress', (data: BenchmarkProgress) => {
      setBenchmarkRunning(true);
      setBenchmarkProgress(data);
    });

    EventsOn('benchmark:complete', (data: TestResultsData) => {
      setResults(data);
      setBenchmarkRunning(false);
      setBenchmarkProgress(null);
    });

    EventsOn('benchmark:error', (_data: { message: string }) => {
      setBenchmarkRunning(false);
      setBenchmarkProgress(null);
    });

    EventsOn('benchmark:stopped', () => {
      setBenchmarkRunning(false);
      setBenchmarkProgress(null);
    });

    return () => {
      EventsOff('benchmark:progress');
      EventsOff('benchmark:complete');
      EventsOff('benchmark:error');
      EventsOff('benchmark:stopped');
    };
  }, [setBenchmarkProgress, setBenchmarkRunning, setResults]);
}
