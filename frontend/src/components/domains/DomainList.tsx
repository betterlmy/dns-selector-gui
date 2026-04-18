import { useState } from 'react';
import { useAppStore } from '../../store/useAppStore';
import {
  RemoveCustomDomain,
  GetDomainList,
} from '../../../wailsjs/go/backend/AppService';
import { AddDomainDialog } from './AddDomainDialog';
import { getErrorMessage } from '../../utils/errors';
import './DomainList.css';

export function DomainList() {
  const domains = useAppStore((s) => s.domains);
  const setDomains = useAppStore((s) => s.setDomains);
  const setError = useAppStore((s) => s.setError);
  const benchmarkRunning = useAppStore((s) => s.benchmarkRunning);
  const [showDialog, setShowDialog] = useState(false);

  const handleRemove = async (domain: string) => {
    try {
      await RemoveCustomDomain(domain);
      const updated = await GetDomainList();
      setDomains(updated);
    } catch (err) {
      setError(getErrorMessage(err, '删除测试域名失败。'));
    }
  };

  const handleAdded = async () => {
    const updated = await GetDomainList();
    setDomains(updated);
    setShowDialog(false);
  };

  return (
    <div className="domain-list">
      <div className="domain-list-header">
        <span className="section-label">测试域名 ({domains.length})</span>
        <button
          className="add-btn"
          onClick={() => setShowDialog(true)}
          disabled={benchmarkRunning}
        >
          + 添加
        </button>
      </div>
      <div className="domain-list-items">
        {domains.map((d) => (
          <div key={d.domain} className="domain-item">
            <span className="domain-name">{d.domain}</span>
            <button
              className="remove-btn"
              onClick={() => handleRemove(d.domain)}
              disabled={d.isPreset || benchmarkRunning}
              title={d.isPreset ? '预设项不可删除' : '删除'}
            >
              ✕
            </button>
          </div>
        ))}
      </div>
      {showDialog && (
        <AddDomainDialog
          onClose={() => setShowDialog(false)}
          onAdded={handleAdded}
        />
      )}
    </div>
  );
}
