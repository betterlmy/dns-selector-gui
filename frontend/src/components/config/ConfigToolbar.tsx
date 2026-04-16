import { useRef, useState } from 'react';
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
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleImportClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileSelected = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setError('');
    setImporting(true);
    try {
      // Wails ImportConfig expects a file path.
      // In a real Wails app, we'd use the Go-side file dialog.
      // For this implementation, we pass the file.name.
      await ImportConfig(file.name);

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
      // Reset file input so same file can be selected again
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  };

  const handleExport = async () => {
    setError('');
    setExporting(true);
    try {
      await ExportConfig('dns-selector-config.json');
    } catch (err: any) {
      setError(typeof err === 'string' ? err : err?.message || '导出配置失败');
    } finally {
      setExporting(false);
    }
  };

  return (
    <div className="config-toolbar">
      <input
        ref={fileInputRef}
        type="file"
        accept=".json"
        style={{ display: 'none' }}
        onChange={handleFileSelected}
      />
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
