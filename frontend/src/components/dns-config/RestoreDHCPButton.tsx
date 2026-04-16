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

  const openDialog = () => {
    setSelected('');
    setFeedback(null);
    setShowDialog(true);
  };

  const closeDialog = () => {
    setShowDialog(false);
  };

  return (
    <>
      <button className="btn-restore-dhcp" onClick={openDialog}>
        恢复默认 DNS
      </button>

      {showDialog && (
        <div className="dialog-overlay" onClick={closeDialog}>
          <div className="dialog" onClick={(e) => e.stopPropagation()}>
            <h3 className="dialog-title">恢复 DHCP 自动获取</h3>

            <div className="form-group">
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
                      name="restore-adapter"
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
              <button className="btn-cancel" onClick={closeDialog} disabled={submitting}>
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
