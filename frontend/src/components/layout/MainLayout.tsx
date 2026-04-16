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

export function MainLayout() {
  return (
    <div className="main-layout">
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
      <main className="content">
        <div className="content-section">
          <BenchmarkControl />
        </div>
        <div className="content-section">
          <Recommendation />
          <ResultsTable />
          <ScoreChart />
        </div>
        <div className="content-section">
          <CurrentDNSDisplay />
          <div style={{ marginTop: 8 }}>
            <RestoreDHCPButton />
          </div>
        </div>
      </main>
    </div>
  );
}
