import { useState } from 'react';
import { useAppStore } from '../../store/useAppStore';
import {
  RemoveCustomServer,
  GetServerList,
} from '../../../wailsjs/go/backend/AppService';
import { AddServerDialog } from './AddServerDialog';
import { getErrorMessage } from '../../utils/errors';
import './ServerList.css';

const protocolLabels: Record<string, string> = {
  udp: 'UDP',
  dot: 'DoT',
  doh: 'DoH',
};

const protocolColors: Record<string, string> = {
  udp: 'var(--accent-color)',
  dot: 'var(--success-color)',
  doh: 'var(--warning-color)',
};

export function ServerList() {
  const servers = useAppStore((s) => s.servers);
  const setServers = useAppStore((s) => s.setServers);
  const setError = useAppStore((s) => s.setError);
  const benchmarkRunning = useAppStore((s) => s.benchmarkRunning);
  const [showDialog, setShowDialog] = useState(false);

  const handleRemove = async (protocol: string, address: string, tlsServerName: string) => {
    try {
      await RemoveCustomServer(protocol, address, tlsServerName);
      const updated = await GetServerList();
      setServers(updated);
    } catch (err) {
      setError(getErrorMessage(err, '删除 DNS 服务器失败。'));
    }
  };

  const handleAdded = async () => {
    const updated = await GetServerList();
    setServers(updated);
    setShowDialog(false);
  };

  return (
    <div className="server-list">
      <div className="server-list-header">
        <span className="section-label">DNS 服务器 ({servers.length})</span>
        <button
          className="add-btn"
          onClick={() => setShowDialog(true)}
          disabled={benchmarkRunning}
        >
          + 添加
        </button>
      </div>
      <div className="server-list-items">
        {servers.map((s) => (
          <div key={`${s.protocol}|${s.address}|${s.tlsServerName}`} className="server-item">
            <span
              className="protocol-badge"
              style={{ backgroundColor: protocolColors[s.protocol] || 'var(--accent-color)' }}
            >
              {protocolLabels[s.protocol] || s.protocol.toUpperCase()}
            </span>
            <div className="server-info">
              <span className="server-name">{s.name}</span>
              <span className="server-address">{s.address}</span>
            </div>
            <button
              className="remove-btn"
              onClick={() => handleRemove(s.protocol, s.address, s.tlsServerName)}
              disabled={s.isPreset || benchmarkRunning}
              title={s.isPreset ? '预设项不可删除' : '删除'}
            >
              ✕
            </button>
          </div>
        ))}
      </div>
      {showDialog && (
        <AddServerDialog
          onClose={() => setShowDialog(false)}
          onAdded={handleAdded}
        />
      )}
    </div>
  );
}
