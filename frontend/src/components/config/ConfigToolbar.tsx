import { useState } from 'react';
import { useAppStore } from '../../store/useAppStore';
import {
  ImportConfig,
  ExportConfig,
  GetServerList,
  GetDomainList,
  GetTestParams,
  GetCurrentPreset,
} from '../../../wailsjs/go/backend/AppService';
import './ConfigToolbar.css';

export function ConfigToolbar() {
  const setServers = useAppStore((s) => s.setServers);
  const setDomains = useAppStore((s) => s.setDomains);
  const setTestParams = useAppStore((s) => s.setTestParams);
  const setCurrentPreset = useAppStore((s) => s.setCurrentPreset);
  const benchmarkRunning = useAppStore((s) => s.benchmarkRunning);

  const [error, setError] = useState('');
  const [importing, setImporting] = useState(false);
  const [exporting, setExporting] = useState(false);

  const handleImportClick = async () => {
    setError('');
    setImporting(true);
    try {
      await ImportConfig('');

      // Refresh all state after import
      const [preset, servers, domains, params] = await Promise.all([
        GetCurrentPreset(),
        GetServerList(),
        GetDomainList(),
        GetTestParams(),
      ]);
      setCurrentPreset(preset);
      setServers(servers);
      setDomains(domains);
      setTestParams(params);
    } catch (err: any) {
      setError(typeof err === 'string' ? err : err?.message || '导入配置失败');
    } finally {
      setImporting(false);
    }
  };

  const handleExport = async () => {
    setError('');
    setExporting(true);
    try {
      await ExportConfig('');
    } catch (err: any) {
      setError(typeof err === 'string' ? err : err?.message || '导出配置失败');
    } finally {
      setExporting(false);
    }
  };

  return (
    <div className="config-toolbar">
      <button
        className="config-btn"
        onClick={handleImportClick}
        disabled={importing || benchmarkRunning}
      >
        {importing ? '导入中...' : '📂 导入配置'}
      </button>
      <button
        className="config-btn"
        onClick={handleExport}
        disabled={exporting || benchmarkRunning}
      >
        {exporting ? '导出中...' : '💾 导出配置'}
      </button>
      {error && (
        <div className="config-error">
          <span>{error}</span>
          <button className="config-error-close" onClick={() => setError('')}>✕</button>
        </div>
      )}
    </div>
  );
}
