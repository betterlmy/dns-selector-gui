import { useState, useRef, useCallback } from 'react';
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

// 列定义：key、标题、默认宽度(px)
const COLUMNS = [
  { key: 'name',       label: '服务器',   defaultWidth: 130 },
  { key: 'address',    label: '地址',     defaultWidth: 160 },
  { key: 'protocol',   label: '协议',     defaultWidth: 56  },
  { key: 'median',     label: '中位延迟', defaultWidth: 80  },
  { key: 'p95',        label: 'P95 延迟', defaultWidth: 80  },
  { key: 'rate',       label: '成功率',   defaultWidth: 68  },
  { key: 'mismatch',   label: '不一致',   defaultWidth: 60  },
  { key: 'score',      label: 'Score',    defaultWidth: 64  },
  { key: 'action',     label: '',         defaultWidth: 60  },
];

export function ResultsTable() {
  const results = useAppStore((s) => s.results);
  const benchmarkRunning = useAppStore((s) => s.benchmarkRunning);
  const [applyTarget, setApplyTarget] = useState<{ address: string; name: string } | null>(null);

  // 列宽状态
  const [colWidths, setColWidths] = useState<number[]>(COLUMNS.map((c) => c.defaultWidth));
  const resizingRef = useRef<{ colIdx: number; startX: number; startWidth: number } | null>(null);

  // 拖拽开始
  const onResizeStart = useCallback((e: React.MouseEvent, colIdx: number) => {
    e.preventDefault();
    resizingRef.current = { colIdx, startX: e.clientX, startWidth: colWidths[colIdx] };

    const onMove = (ev: MouseEvent) => {
      if (!resizingRef.current) return;
      const { colIdx: ci, startX, startWidth } = resizingRef.current;
      const delta = ev.clientX - startX;
      const newWidth = Math.max(40, startWidth + delta);
      setColWidths((prev) => {
        const next = [...prev];
        next[ci] = newWidth;
        return next;
      });
    };
    const onUp = () => {
      resizingRef.current = null;
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('mouseup', onUp);
    };
    window.addEventListener('mousemove', onMove);
    window.addEventListener('mouseup', onUp);
  }, [colWidths]);

  if (!results || !results.items || results.items.length === 0) {
    return (
      <div className="results-table-wrapper">
        <div className="results-empty">尚未进行测试</div>
      </div>
    );
  }

  const items = results.items;
  const bestScore = items.length > 0 ? items[0].score : 0;

  const rowClass = (item: TestResultItem, idx: number) => {
    if (item.isTimeout || item.score === 0) return 'row--timeout';
    if (idx === 0 && bestScore > 0) return 'row--best';
    return '';
  };
  const scoreClass = (item: TestResultItem, idx: number) => {
    if (item.isTimeout || item.score === 0) return 'score--zero';
    if (idx === 0 && bestScore > 0) return 'score--best';
    return '';
  };
  const canApply = (item: TestResultItem) =>
    !item.isTimeout && item.score > 0 && !benchmarkRunning && item.canApplyToSystem;
  const applyTitle = (item: TestResultItem) => {
    if (!item.canApplyToSystem) return '该服务器可测速，但无法直接写入系统 DNS（需配置 BootstrapIP）';
    if (item.isTimeout || item.score === 0) return undefined;
    return '双击应用为系统 DNS';
  };
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

      <div className="results-table-scroll">
        <table className="results-table" style={{ width: colWidths.reduce((a, b) => a + b, 0) }}>
          <colgroup>
            {colWidths.map((w, i) => <col key={i} style={{ width: w }} />)}
          </colgroup>
          <thead>
            <tr>
              {COLUMNS.map((col, i) => (
                <th key={col.key} style={{ width: colWidths[i], minWidth: colWidths[i] }}>
                  <div className="th-inner">
                    <span>{col.label}</span>
                    {/* 最后一列（操作列）不加 resize handle */}
                    {i < COLUMNS.length - 1 && (
                      <span
                        className="col-resize-handle"
                        onMouseDown={(e) => onResizeStart(e, i)}
                      />
                    )}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {items.map((item, idx) => (
              <tr
                key={`${item.address}-${item.protocol}`}
                className={`${rowClass(item, idx)} result-row${canApply(item) ? ' result-row--clickable' : ''}`}
                onDoubleClick={() => handleApply(item)}
                title={applyTitle(item)}
              >
                <td title={item.name}>{item.name}</td>
                <td className="addr-cell" title={item.address}>{item.address}</td>
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
                  ) : item.answerMismatches}
                </td>
                <td className={scoreClass(item, idx)}>{formatScore(item.score)}</td>
                <td className="action-cell">
                  <button
                    className="apply-dns-btn"
                    onClick={(e) => { e.stopPropagation(); handleApply(item); }}
                    disabled={!canApply(item)}
                    title={applyTitle(item)}
                    tabIndex={-1}
                  >
                    应用
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

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
