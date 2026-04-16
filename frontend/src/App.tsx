import { useEffect } from 'react';
import { useAppStore } from './store/useAppStore';
import { applyTheme } from './styles/theme';
import { useWailsEvents } from './hooks/useWailsEvents';
import { MainLayout } from './components/layout/MainLayout';

function App() {
  const theme = useAppStore((s) => s.theme);
  const initialize = useAppStore((s) => s.initialize);

  // 主题变化时应用到 DOM
  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  // 应用启动时从后端加载数据
  useEffect(() => {
    initialize();
  }, [initialize]);

  // 监听 Wails 后端事件
  useWailsEvents();

  return <MainLayout />;
}

export default App;
