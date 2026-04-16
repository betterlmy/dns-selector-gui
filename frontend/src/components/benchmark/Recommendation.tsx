import { useState } from 'react';
import { useAppStore } from '../../store/useAppStore';
import { ApplyDNSDialog } from '../dns-config/ApplyDNSDialog';
import './Recommendation.css';

export function Recommendation() {
  const results = useAppStore((s) => s.results);
  const benchmarkRunning = useAppStore((s) => s.benchmarkRunning);
  const [showApply, setShowApply] = useState(false);

  if (!results || !results.bestDNS || !results.items || results.items.length === 0) {
    return null;
  }

  const best = results.items[0];
  if (!best || best.score === 0) return null;

  return (
    <>
      <div className="recommendation">
        <span className="recommendation__icon">🏆</span>
        <div className="recommendation__body">
          <span className="recommendation__title">推荐 DNS</span>
          <span className="recommendation__name">{best.name}</span>
          <span className="recommendation__detail">
            {best.address} · {best.protocol.toUpperCase()}
          </span>
        </div>
        <span className="recommendation__score">{best.score.toFixed(1)}</span>
        <button
          className="recommendation__apply-btn"
          onClick={() => setShowApply(true)}
          disabled={benchmarkRunning}
          title={`将 ${best.name} 设为系统 DNS`}
        >
          应用
        </button>
      </div>

      {showApply && (
        <ApplyDNSDialog
          dnsAddress={best.address}
          dnsName={best.name}
          onClose={() => setShowApply(false)}
        />
      )}
    </>
  );
}
