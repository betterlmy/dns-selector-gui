import { useEffect, useState } from 'react';
import { useAppStore } from '../../store/useAppStore';
import { GetNetworkAdapters } from '../../../wailsjs/go/backend/AppService';
import { getErrorMessage } from '../../utils/errors';
import './DnsConfig.css';

export function CurrentDNSDisplay() {
  const adapters = useAppStore((s) => s.adapters);
  const setAdapters = useAppStore((s) => s.setAdapters);
  const setError = useAppStore((s) => s.setError);
  const [loadError, setLoadError] = useState<string | null>(null);

  const loadAdapters = async () => {
    try {
      const list = await GetNetworkAdapters();
      setAdapters(list ?? []);
      setLoadError(null);
    } catch (err) {
      const message = getErrorMessage(err, '读取网络适配器失败。');
      setLoadError(message);
      setError(message);
      setAdapters([]);
    }
  };

  // 组件挂载时自动加载
  useEffect(() => {
    if (adapters.length === 0) {
      loadAdapters();
    }
  }, []);

  return (
    <div className="dns-display">
      <div className="dns-display-header">
        <span className="section-label">当前 DNS 配置</span>
        <button className="btn-refresh" onClick={loadAdapters}>
          刷新
        </button>
      </div>

      {loadError ? (
        <div className="dns-empty dns-empty--error">{loadError}</div>
      ) : (!adapters || adapters.length === 0) ? (
        <div className="dns-empty">无活动网络适配器</div>
      ) : (
        adapters.map((a) => (
          <div className="adapter-card" key={a.name}>
            <div className="adapter-card-name">{a.name}</div>
            {a.ipAddresses && a.ipAddresses.length > 0 && (
              <div className="adapter-card-status">
                IP：{a.ipAddresses.join('，')}
              </div>
            )}
            <div className="adapter-card-dns">
              {a.currentDNS && a.currentDNS.length > 0 ? (
                a.currentDNS.map((dns) => (
                  <span className="dns-tag" key={dns}>{dns}</span>
                ))
              ) : (
                <span className="dns-tag dns-tag--empty">DHCP (自动获取)</span>
              )}
            </div>
          </div>
        ))
      )}
    </div>
  );
}
