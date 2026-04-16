import { useState } from 'react';
import { useAppStore } from '../../store/useAppStore';
import {
  ApplyDNS,
  GetNetworkAdapters,
} from '../../../wailsjs/go/backend/AppService';
import './DnsConfig.css';

interface Props {
  dnsAddress: string;
  dnsName: string;
  onClose: () => void;
}

export function ApplyDNSDialog({ dnsAddress, dnsName, onClose }: Props) {
  const adapters = useAppStore((s) => s.adapters);
  const isAdmin = useAppStore((s) => s.isAdmin);
  const setAdapters = useAppStore((s) => s.setAdapters);

  const [selected, setSelected] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; msg: string } | null>(null);

  const handleApply = async () => {
    if (!selected) return;
    setFeedback(null);
    setSubmitting(true);
    try {
      await ApplyDNS(selected, dnsAddress);
      setFeedback({ type: 'success', msg: `已将 ${dnsName} 应用到 ${selected}` });
      // refresh adapters
      try {
        const list = await GetNetworkAdapters();
        setAdapters(list ?? []);
      } catch { /* ignore */ }
    } catch (err: any) {
      setFeedback({
        type: 'error',
        msg: typeof err === 'string' ? err : err?.message || 'DNS 设置失败',
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="dialog-overlay" onClick={onClose}>
      <div className="dialog" onClick={(e) => e.stopPropagation()}>
        <h3 className="dialog-title">应用 DNS: {dnsName}</h3>

        {!isAdmin && (
          <div className="admin-warning">
            ⚠️ 当前未以管理员身份运行，修改 DNS 可能失败。请以管理员身份重新运行应用。
          </div>
        )}

        <div className="form-group" style={{ marginTop: 12 }}>
          <label className="form-label">选择网络适配器</label>
          {(!adapters || adapters.length === 0) ? (
            <div className="dns-empty">无可用网络适配器</div>
          ) : (
            adapters.map((a) => (
              <label
                key={a.name}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  padding: '6px 0',
                  cursor: 'pointer',
                  fontSize: 13,
                }}
              >
                <input
                  type="radio"
                  name="adapter"
                  value={a.name}
                  checked={selected === a.name}
                  onChange={() => setSelected(a.name)}
                />
                <span>{a.name}</span>
                <span style={{ color: 'var(--text-secondary)', fontSize: 11 }}>
                  ({a.currentDNS?.join(', ') || 'DHCP'})
                </span>
              </label>
            ))
          )}
        </div>

        {feedback && (
          <div className={`dns-feedback dns-feedback--${feedback.type}`}>
            {feedback.msg}
          </div>
        )}

        <div className="dialog-actions">
          <button className="btn-cancel" onClick={onClose} disabled={submitting}>
            {feedback?.type === 'success' ? '关闭' : '取消'}
          </button>
          {feedback?.type !== 'success' && (
            <button
              className="btn-submit"
              onClick={handleApply}
              disabled={submitting || !selected}
            >
              {submitting ? '应用中...' : '应用'}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
