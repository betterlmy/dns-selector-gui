import { useState } from 'react';
import { useAppStore } from '../../store/useAppStore';
import {
  RestoreDHCP,
  GetNetworkAdapters,
} from '../../../wailsjs/go/backend/AppService';
import './DnsConfig.css';

export function RestoreDHCPButton() {
  const adapters = useAppStore((s) => s.adapters);
  const setAdapters = useAppStore((s) => s.setAdapters);

  const [showDialog, setShowDialog] = useState(false);
  const [selected, setSelected] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; msg: string } | null>(null);

  const handleRestore = async () => {
    if (!selected) return;
    setFeedback(null);
    setSubmitting(true);
    try {
      await RestoreDHCP(selected);
      setFeedback({ type: 'success', msg: `已恢复 ${selected} 为 DHCP 自动获取` });
      try {
        const list = await GetNetworkAdapters();
        setAdapters(list ?? []);
      } catch { /* ignore */ }
    } catch (err: any) {
      setFeedback({
        type: 'error',
        msg: typeof err === 'string' ? err : err?.message || '恢复 DHCP 失败',
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <button className="btn-restore-dhcp" onClick={() => { setSelected(''); setFeedback(null); setShowDialog(true); }}>
        恢复默认 DNS
      </button>

      {showDialog && (
        <div className="dialog-overlay" onClick={() => setShowDialog(false)}>
          <div className="dialog" onClick={(e) => e.stopPropagation()}>
            <h3 className="dialog-title">恢复 DHCP 自动获取</h3>

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
                        name="restore-adapter"
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
              <button className="btn-cancel" onClick={() => setShowDialog(false)} disabled={submitting}>
                {feedback?.type === 'success' ? '关闭' : '取消'}
              </button>
              {feedback?.type !== 'success' && (
                <button
                  className="btn-submit"
                  onClick={handleRestore}
                  disabled={submitting || !selected}
                >
                  {submitting ? '恢复中...' : '恢复'}
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
