import { useState } from 'react';
import { AddCustomServer } from '../../../wailsjs/go/backend/AppService';
import './AddServerDialog.css';

interface Props {
  onClose: () => void;
  onAdded: () => void;
}

const protocols = [
  { value: 'udp', label: 'UDP', hint: '例: 8.8.8.8' },
  { value: 'dot', label: 'DoT', hint: '例: dns.google 或 IP@TLSServerName' },
  { value: 'doh', label: 'DoH', hint: '例: https://dns.google/dns-query' },
];

export function AddServerDialog({ onClose, onAdded }: Props) {
  const [protocol, setProtocol] = useState('udp');
  const [address, setAddress] = useState('');
  const [tlsServerName, setTlsServerName] = useState('');
  const [bootstrapIP, setBootstrapIP] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const showTlsFields = protocol === 'dot' || protocol === 'doh';
  const currentHint = protocols.find((p) => p.value === protocol)?.hint || '';

  const handleSubmit = async () => {
    if (!address.trim()) {
      setError('请输入服务器地址');
      return;
    }
    setError('');
    setSubmitting(true);
    try {
      await AddCustomServer({
        protocol,
        address: address.trim(),
        tlsServerName: tlsServerName.trim(),
        bootstrapIP: bootstrapIP.trim(),
      });
      onAdded();
    } catch (err: any) {
      setError(typeof err === 'string' ? err : err?.message || '添加失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="dialog-overlay" onClick={onClose}>
      <div className="dialog" onClick={(e) => e.stopPropagation()}>
        <h3 className="dialog-title">添加 DNS 服务器</h3>

        <div className="form-group">
          <label className="form-label">协议</label>
          <div className="protocol-select">
            {protocols.map((p) => (
              <button
                key={p.value}
                className={`protocol-option ${protocol === p.value ? 'active' : ''}`}
                onClick={() => { setProtocol(p.value); setError(''); }}
              >
                {p.label}
              </button>
            ))}
          </div>
        </div>

        <div className="form-group">
          <label className="form-label">地址</label>
          <input
            className="form-input"
            type="text"
            value={address}
            onChange={(e) => { setAddress(e.target.value); setError(''); }}
            placeholder={currentHint}
            autoFocus
          />
        </div>

        {showTlsFields && (
          <>
            <div className="form-group">
              <label className="form-label">TLS Server Name (可选)</label>
              <input
                className="form-input"
                type="text"
                value={tlsServerName}
                onChange={(e) => setTlsServerName(e.target.value)}
                placeholder="例: dns.google"
              />
            </div>
            <div className="form-group">
              <label className="form-label">Bootstrap IP (可选)</label>
              <input
                className="form-input"
                type="text"
                value={bootstrapIP}
                onChange={(e) => setBootstrapIP(e.target.value)}
                placeholder="例: 8.8.8.8"
              />
            </div>
          </>
        )}

        {error && <div className="form-error">{error}</div>}

        <div className="dialog-actions">
          <button className="btn-cancel" onClick={onClose} disabled={submitting}>
            取消
          </button>
          <button className="btn-submit" onClick={handleSubmit} disabled={submitting}>
            {submitting ? '添加中...' : '添加'}
          </button>
        </div>
      </div>
    </div>
  );
}
