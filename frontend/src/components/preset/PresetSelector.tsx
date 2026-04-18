import { useState } from 'react';
import { useAppStore } from '../../store/useAppStore';
import {
  SwitchPreset,
  GetServerList,
  GetDomainList,
} from '../../../wailsjs/go/backend/AppService';
import { getErrorMessage } from '../../utils/errors';
import './PresetSelector.css';

export function PresetSelector() {
  const currentPreset = useAppStore((s) => s.currentPreset);
  const setCurrentPreset = useAppStore((s) => s.setCurrentPreset);
  const setServers = useAppStore((s) => s.setServers);
  const setDomains = useAppStore((s) => s.setDomains);
  const setError = useAppStore((s) => s.setError);
  const benchmarkRunning = useAppStore((s) => s.benchmarkRunning);
  const [switching, setSwitching] = useState(false);

  const handleSwitch = async (preset: string) => {
    if (preset === currentPreset || switching || benchmarkRunning) return;
    setSwitching(true);
    try {
      await SwitchPreset(preset);
      setCurrentPreset(preset);
      const [servers, domains] = await Promise.all([
        GetServerList(),
        GetDomainList(),
      ]);
      setServers(servers);
      setDomains(domains);
    } catch (err) {
      setError(getErrorMessage(err, '切换预设失败。'));
    } finally {
      setSwitching(false);
    }
  };

  return (
    <div className="preset-selector">
      <span className="preset-selector-label">预设方案</span>
      <div className="preset-buttons">
        <button
          className={`preset-btn ${currentPreset === 'cn' ? 'active' : ''}`}
          onClick={() => handleSwitch('cn')}
          disabled={switching || benchmarkRunning}
        >
          🇨🇳 CN
        </button>
        <button
          className={`preset-btn ${currentPreset === 'global' ? 'active' : ''}`}
          onClick={() => handleSwitch('global')}
          disabled={switching || benchmarkRunning}
        >
          🌍 Global
        </button>
      </div>
    </div>
  );
}
