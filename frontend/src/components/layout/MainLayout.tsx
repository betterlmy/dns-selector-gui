import { useState } from 'react';
import { PresetSelector } from '../preset/PresetSelector';
import { ServerList } from '../servers/ServerList';
import { DomainList } from '../domains/DomainList';
import { TestParamsForm } from '../params/TestParamsForm';
import { ConfigToolbar } from '../config/ConfigToolbar';
import { BenchmarkControl } from '../benchmark/BenchmarkControl';
import { Recommendation } from '../benchmark/Recommendation';
import { ResultsTable } from '../benchmark/ResultsTable';
import { ScoreChart } from '../benchmark/ScoreChart';
import { CurrentDNSDisplay } from '../dns-config/CurrentDNSDisplay';
import { RestoreDHCPButton } from '../dns-config/RestoreDHCPButton';
import './MainLayout.css';

type Tab = 'results' | 'dns';

export function MainLayout() {
  const [activeTab, setActiveTab] = useState<Tab>('results');
  const [sidebarOpen, setSidebarOpen] = useState(true);

  return (
    <div className="main-layout">
      {/* 侧边栏折叠按钮 */}
      <button
        className="sidebar-toggle"
        onClick={() => setSidebarOpen(!sidebarOpen)}
        title={sidebarOpen ? '收起配置面板' : '展开配置面板'}
      >
        {sidebarOpen ? '◀' : '▶'}
      </button>

      {/* 左侧配置栏 */}
      {sidebarOpen && (
        <aside className="sidebar">
          <div className="sidebar-section">
            <PresetSelector />
          </div>
          <div className="sidebar-section">
            <ServerList />
          </div>
          <div className="sidebar-section">
            <DomainList />
          </div>
          <div className="sidebar-section">
            <TestParamsForm />
          </div>
          <div className="sidebar-section">
            <ConfigToolbar />
          </div>
        </aside>
      )}

      {/* 右侧主区域 */}
      <main className="content">
        <div className="content-section">
          <BenchmarkControl />
        </div>

        <div className="tab-bar">
          <button
            className={`tab-btn${activeTab === 'results' ? ' tab-btn--active' : ''}`}
            onClick={() => setActiveTab('results')}
          >
            测试结果
          </button>
          <button
            className={`tab-btn${activeTab === 'dns' ? ' tab-btn--active' : ''}`}
            onClick={() => setActiveTab('dns')}
          >
            DNS 配置
          </button>
        </div>

        <div className="tab-content">
          {activeTab === 'results' && (
            <div className="content-section">
              <Recommendation />
              <ResultsTable />
              <ScoreChart />
            </div>
          )}

          {activeTab === 'dns' && (
            <div className="content-section">
              <CurrentDNSDisplay />
              <div style={{ marginTop: 8 }}>
                <RestoreDHCPButton />
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
