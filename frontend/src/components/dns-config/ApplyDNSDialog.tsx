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
  const [secondary, setSecondary] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; msg: string } | null>(null);

  const handleApply = async () => {
    if (!selected) return;
    setFeedback(null);
    setSubmitting(true);
    try {
      await ApplyDNS(selected, dnsAddress, secondary.trim());
      const secMsg = secondary.trim() ? `，备用 ${secondary.trim()}` : '';
      setFeedback({ type: 'success', msg: `已将首选 DNS ${dnsName}${secMsg} 应用到 ${selected}` });
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
        <h3 className="dialog-title">应用 DNS</h3>

        {/* 首选 DNS */}
        <div className="dns-role-row">
          <span className="dns-role-badge dns-role-badge--primary">首选</span>
          <div className="dns-role-info">
            <span className="dns-target-value">{dnsName}</span>
            <span className="dns-target-addr">{dnsAddress}</span>
          </div>
        </div>

        {/* 备用 DNS 输入 */}
        <div className="dns-role-row">
          <span className="dns-role-badge dns-role-badge--secondary">备用</span>
          <input
            className="form-input dns-secondary-input"
            type="text"
            placeholder="输入备用 DNS IP（可选，如 8.8.4.4）"
            value={secondary}
            onChange={(e) => setSecondary(e.target.value)}
            disabled={submitting}
          />
        </div>

        {!isAdmin && (
          <div className="admin-warning">
            ⚠️ 当前未以管理员身份运行，修改 DNS 可能失败。请以管理员身份重新运行应用。
          </div>
        )}

        <div className="form-group">
          <label className="form-label">选择网络适配器</label>
          {(!adapters || adapters.length === 0) ? (
            <div className="dns-empty">无可用网络适配器</div>
          ) : (
            <div className="adapter-list">
              {adapters.map((a) => (
                <label
                  key={a.name}
                  className={`adapter-option${selected === a.name ? ' adapter-option--selected' : ''}`}
                >
                  <input
                    type="radio"
                    name="adapter"
                    value={a.name}
                    checked={selected === a.name}
                    onChange={() => setSelected(a.name)}
                  />
                  <div className="adapter-option__info">
                    <span className="adapter-option__name">{a.name}</span>
                    {a.ipAddresses?.length > 0 && (
                      <span className="adapter-option__ip">
                        IP：{a.ipAddresses.join('，')}
                      </span>
                    )}
                    <span className="adapter-option__dns">
                      当前 DNS：{a.currentDNS?.length ? a.currentDNS.join('，') : 'DHCP 自动获取'}
                    </span>
                  </div>
                </label>
              ))}
            </div>
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
