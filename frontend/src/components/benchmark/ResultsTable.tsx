import { useState } from 'react';
import { useAppStore } from '../../store/useAppStore';
import type { TestResultItem } from '../../types';
import { ApplyDNSDialog } from '../dns-config/ApplyDNSDialog';
import './ResultsTable.css';

function formatLatency(ms: number): string {
  return ms > 0 ? `${ms.toFixed(1)} ms` : '—';
}

function formatRate(rate: number): string {
  return `${(rate * 100).toFixed(1)}%`;
}

function formatScore(score: number): string {
  return score.toFixed(1);
}

function protocolClass(protocol: string): string {
  const p = protocol.toLowerCase();
  if (p === 'dot') return 'protocol-tag protocol-tag--dot';
  if (p === 'doh') return 'protocol-tag protocol-tag--doh';
  return 'protocol-tag protocol-tag--udp';
}

export function ResultsTable() {
  const results = useAppStore((s) => s.results);
  const benchmarkRunning = useAppStore((s) => s.benchmarkRunning);
  const [applyTarget, setApplyTarget] = useState<{ address: string; name: string } | null>(null);

  if (!results || !results.items || results.items.length === 0) {
    return (
      <div className="results-table-wrapper">
        <div className="results-empty">尚未进行测试</div>
      </div>
    );
  }

  const items = results.items;
  const bestScore = items.length > 0 ? items[0].score : 0;

  const rowClass = (item: TestResultItem, idx: number): string => {
    if (item.isTimeout || item.score === 0) return 'row--timeout';
    if (idx === 0 && bestScore > 0) return 'row--best';
    return '';
  };

  const scoreClass = (item: TestResultItem, idx: number): string => {
    if (item.isTimeout || item.score === 0) return 'score--zero';
    if (idx === 0 && bestScore > 0) return 'score--best';
    return '';
  };

  const canApply = (item: TestResultItem) =>
    !item.isTimeout && item.score > 0 && !benchmarkRunning;

  const handleApply = (item: TestResultItem) => {
    if (!canApply(item)) return;
    setApplyTarget({ address: item.address, name: item.name });
  };

  return (
    <div className="results-table-wrapper">
      <div className="results-table-header">
        <span className="section-label">测试结果</span>
        {results.testTime && (
          <span className="results-table-time">
            测试时间: {new Date(results.testTime).toLocaleString()}
          </span>
        )}
      </div>
      <table className="results-table">
        <thead>
          <tr>
            <th>服务器</th>
            <th>地址</th>
            <th>协议</th>
            <th>中位延迟</th>
            <th>P95 延迟</th>
            <th>成功率</th>
            <th>不一致</th>
            <th>Score</th>
            {/* 操作列宽度固定，避免 hover 时抖动 */}
            <th className="action-th"></th>
          </tr>
        </thead>
        <tbody>
          {items.map((item, idx) => (
            <tr
              key={`${item.address}-${item.protocol}`}
              className={`${rowClass(item, idx)} result-row${canApply(item) ? ' result-row--clickable' : ''}`}
              onDoubleClick={() => handleApply(item)}
              title={canApply(item) ? '双击应用为系统 DNS' : undefined}
            >
              <td>{item.name}</td>
              <td className="addr-cell">{item.address}</td>
              <td>
                <span className={protocolClass(item.protocol)}>
                  {item.protocol.toUpperCase()}
                </span>
              </td>
              <td>{formatLatency(item.medianLatencyMs)}</td>
              <td>{formatLatency(item.p95LatencyMs)}</td>
              <td>{formatRate(item.successRate)}</td>
              <td>
                {item.answerMismatches > 0 ? (
                  <span className="mismatch-warn" title="答案与多数服务器不一致">
                    ⚠️ {item.answerMismatches}
                  </span>
                ) : (
                  item.answerMismatches
                )}
              </td>
              <td className={scoreClass(item, idx)}>{formatScore(item.score)}</td>
              {/* 操作列：visibility 由 CSS hover 控制，列宽始终固定 */}
              <td className="action-cell">
                <button
                  className="apply-dns-btn"
                  onClick={(e) => { e.stopPropagation(); handleApply(item); }}
                  tabIndex={-1}
                >
                  应用
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {applyTarget && (
        <ApplyDNSDialog
          dnsAddress={applyTarget.address}
          dnsName={applyTarget.name}
          onClose={() => setApplyTarget(null)}
        />
      )}
    </div>
  );
}
