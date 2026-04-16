import { useAppStore } from '../../store/useAppStore';
import { StartBenchmark, StopBenchmark } from '../../../wailsjs/go/backend/AppService';
import './BenchmarkControl.css';

export function BenchmarkControl() {
  const benchmarkRunning = useAppStore((s) => s.benchmarkRunning);
  const benchmarkProgress = useAppStore((s) => s.benchmarkProgress);
  const setBenchmarkRunning = useAppStore((s) => s.setBenchmarkRunning);

  const handleStart = async () => {
    setBenchmarkRunning(true);
    try {
      await StartBenchmark();
    } catch {
      setBenchmarkRunning(false);
    }
  };

  const handleStop = async () => {
    try {
      await StopBenchmark();
    } catch {
      // stop may fail if already stopped
    }
  };

  return (
    <div className="benchmark-control">
      <div className="benchmark-control-header">
        <button
          className={`benchmark-btn ${benchmarkRunning ? 'benchmark-btn--stop' : 'benchmark-btn--start'}`}
          onClick={benchmarkRunning ? handleStop : handleStart}
        >
          {benchmarkRunning ? '⏹ 停止测试' : '▶ 开始测试'}
        </button>
        {benchmarkProgress && (
          <div className="progress-bar">
            <div
              className="progress-bar__fill"
              style={{ width: `${benchmarkProgress.percent}%` }}
            />
            <span className="progress-bar__text">
              {benchmarkProgress.completed}/{benchmarkProgress.total} ({Math.round(benchmarkProgress.percent)}%)
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
