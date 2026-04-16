import { useState } from 'react';
import { AddCustomDomain } from '../../../wailsjs/go/backend/AppService';
import '../servers/AddServerDialog.css';

interface Props {
  onClose: () => void;
  onAdded: () => void;
}

export function AddDomainDialog({ onClose, onAdded }: Props) {
  const [domain, setDomain] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (!domain.trim()) {
      setError('请输入域名');
      return;
    }
    setError('');
    setSubmitting(true);
    try {
      await AddCustomDomain(domain.trim());
      onAdded();
    } catch (err: any) {
      setError(typeof err === 'string' ? err : err?.message || '添加失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !submitting) handleSubmit();
  };

  return (
    <div className="dialog-overlay" onClick={onClose}>
      <div className="dialog" onClick={(e) => e.stopPropagation()}>
        <h3 className="dialog-title">添加测试域名</h3>

        <div className="form-group">
          <label className="form-label">域名</label>
          <input
            className="form-input"
            type="text"
            value={domain}
            onChange={(e) => { setDomain(e.target.value); setError(''); }}
            onKeyDown={handleKeyDown}
            placeholder="例: example.com"
            autoFocus
          />
        </div>

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
