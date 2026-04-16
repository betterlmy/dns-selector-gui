import { useAppStore } from '../../store/useAppStore';
import { GetNetworkAdapters } from '../../../wailsjs/go/backend/AppService';
import './DnsConfig.css';

export function CurrentDNSDisplay() {
  const adapters = useAppStore((s) => s.adapters);
  const setAdapters = useAppStore((s) => s.setAdapters);

  const handleRefresh = async () => {
    try {
      const list = await GetNetworkAdapters();
      setAdapters(list ?? []);
    } catch {
      // silently ignore
    }
  };

  return (
    <div className="dns-display">
      <div className="dns-display-header">
        <span className="section-label">当前 DNS 配置</span>
        <button className="btn-refresh" onClick={handleRefresh}>
          刷新
        </button>
      </div>

      {(!adapters || adapters.length === 0) ? (
        <div className="dns-empty">无活动网络适配器</div>
      ) : (
        adapters.map((a) => (
          <div className="adapter-card" key={a.name}>
            <div className="adapter-card-name">{a.name}</div>
            <div className="adapter-card-status">状态: {a.status}</div>
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
