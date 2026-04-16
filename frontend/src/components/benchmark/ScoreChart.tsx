import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts';
import { useAppStore } from '../../store/useAppStore';
import type { TestResultItem } from '../../types';
import './ScoreChart.css';

interface ChartItem {
  name: string;
  score: number;
  item: TestResultItem;
}

function CustomTooltip({ active, payload }: { active?: boolean; payload?: Array<{ payload: ChartItem }> }) {
  if (!active || !payload || payload.length === 0) return null;
  const data = payload[0].payload;
  const item = data.item;
  return (
    <div className="score-tooltip">
      <div className="score-tooltip__name">{item.name} ({item.protocol.toUpperCase()})</div>
      <div className="score-tooltip__row">中位延迟: <span>{item.medianLatencyMs.toFixed(1)} ms</span></div>
      <div className="score-tooltip__row">P95 延迟: <span>{item.p95LatencyMs.toFixed(1)} ms</span></div>
      <div className="score-tooltip__row">成功率: <span>{(item.successRate * 100).toFixed(1)}%</span></div>
      <div className="score-tooltip__row">不一致: <span>{item.answerMismatches}</span></div>
      <div className="score-tooltip__row">Score: <span>{item.score.toFixed(1)}</span></div>
    </div>
  );
}

export function ScoreChart() {
  const results = useAppStore((s) => s.results);

  if (!results || !results.items || results.items.length === 0) {
    return null;
  }

  const items = results.items;
  const bestScore = items.length > 0 ? items[0].score : 0;

  const chartData: ChartItem[] = items.map((item) => ({
    name: item.name,
    score: parseFloat(item.score.toFixed(1)),
    item,
  }));

  const barColor = (item: TestResultItem): string => {
    if (item.isTimeout || item.score === 0) return 'var(--score-zero)';
    if (item.score === bestScore && bestScore > 0) return 'var(--score-best)';
    return 'var(--accent-color)';
  };

  return (
    <div className="score-chart-wrapper">
      <span className="section-label">Score 对比</span>
      <div className="score-chart-container">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={chartData} margin={{ top: 8, right: 16, bottom: 4, left: 0 }}>
            <XAxis
              dataKey="name"
              tick={{ fontSize: 11, fill: 'var(--text-secondary)' }}
              angle={-35}
              textAnchor="end"
              height={60}
              interval={0}
            />
            <YAxis tick={{ fontSize: 11, fill: 'var(--text-secondary)' }} />
            <Tooltip content={<CustomTooltip />} cursor={{ fill: 'var(--bg-tertiary)' }} />
            <Bar dataKey="score" radius={[4, 4, 0, 0]}>
              {chartData.map((entry, idx) => (
                <Cell key={idx} fill={barColor(entry.item)} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
