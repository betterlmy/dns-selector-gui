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

  const isValidIPv4 = (value: string) => {
    const parts = value.split('.');
    if (parts.length !== 4) return false;
    return parts.every((part) => /^\d+$/.test(part) && Number(part) >= 0 && Number(part) <= 255);
  };

  const isValidDomain = (value: string) => /^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$/.test(value);

  const isValidHTTPSURL = (value: string) => {
    try {
      const parsed = new URL(value);
      return parsed.protocol === 'https:' && parsed.host.length > 0;
    } catch {
      return false;
    }
  };

  const validateForm = () => {
    const trimmedAddress = address.trim();
    const trimmedBootstrapIP = bootstrapIP.trim();
    const trimmedTLSServerName = tlsServerName.trim();

    if (!trimmedAddress) {
      return '请输入服务器地址';
    }

    if (protocol === 'udp' && !isValidIPv4(trimmedAddress)) {
      return 'UDP 地址必须是合法的 IPv4 地址';
    }

    if (protocol === 'dot') {
      const parts = trimmedAddress.split('@');
      if (parts.length === 2) {
        if (!isValidIPv4(parts[0]) || !parts[1]) {
          return 'DoT 地址必须是合法域名或 IP@TLSServerName';
        }
      } else if (!isValidDomain(trimmedAddress)) {
        return 'DoT 地址必须是合法域名或 IP@TLSServerName';
      }
    }

    if (protocol === 'doh') {
      const atIndex = trimmedAddress.lastIndexOf('@');
      if (atIndex !== -1) {
        const urlPart = trimmedAddress.slice(0, atIndex);
        const tlsPart = trimmedAddress.slice(atIndex + 1);
        if (!isValidHTTPSURL(urlPart) || !tlsPart) {
          return 'DoH 地址必须是合法的 HTTPS URL';
        }
      } else if (!isValidHTTPSURL(trimmedAddress)) {
        return 'DoH 地址必须是合法的 HTTPS URL';
      }
    }

    if (trimmedBootstrapIP && !isValidIPv4(trimmedBootstrapIP)) {
      return 'Bootstrap IP 必须是合法的 IPv4 地址';
    }

    if (trimmedTLSServerName && !isValidDomain(trimmedTLSServerName)) {
      return 'TLS Server Name 必须是合法域名';
    }

    return '';
  };

  const validationError = validateForm();

  const handleSubmit = async () => {
    if (validationError) {
      setError(validationError);
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
            className={`form-input ${error || validationError ? 'has-error' : ''}`}
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
                className={`form-input ${error || validationError ? 'has-error' : ''}`}
                type="text"
                value={tlsServerName}
                onChange={(e) => { setTlsServerName(e.target.value); setError(''); }}
                placeholder="例: dns.google"
              />
            </div>
            <div className="form-group">
              <label className="form-label">Bootstrap IP (可选)</label>
              <input
                className={`form-input ${error || validationError ? 'has-error' : ''}`}
                type="text"
                value={bootstrapIP}
                onChange={(e) => { setBootstrapIP(e.target.value); setError(''); }}
                placeholder="例: 8.8.8.8"
              />
            </div>
          </>
        )}

        {(error || validationError) && <div className="form-error">{error || validationError}</div>}

        <div className="dialog-actions">
          <button className="btn-cancel" onClick={onClose} disabled={submitting}>
            取消
          </button>
          <button className="btn-submit" onClick={handleSubmit} disabled={submitting || !!validationError}>
            {submitting ? '添加中...' : '添加'}
          </button>
        </div>
      </div>
    </div>
  );
}
