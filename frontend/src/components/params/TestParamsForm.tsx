import { useState, useEffect } from 'react';
import { useAppStore } from '../../store/useAppStore';
import { SetTestParams } from '../../../wailsjs/go/backend/AppService';
import type { TestParams } from '../../types';
import { getErrorMessage } from '../../utils/errors';
import './TestParamsForm.css';

interface FieldError {
  queries?: string;
  warmup?: string;
  concurrency?: string;
  timeout?: string;
}

export function TestParamsForm() {
  const testParams = useAppStore((s) => s.testParams);
  const setTestParams = useAppStore((s) => s.setTestParams);
  const benchmarkRunning = useAppStore((s) => s.benchmarkRunning);
  const setError = useAppStore((s) => s.setError);

  const [form, setForm] = useState({
    queries: String(testParams.queries),
    warmup: String(testParams.warmup),
    concurrency: String(testParams.concurrency),
    timeout: String(testParams.timeout),
  });
  const [errors, setErrors] = useState<FieldError>({});

  useEffect(() => {
    setForm({
      queries: String(testParams.queries),
      warmup: String(testParams.warmup),
      concurrency: String(testParams.concurrency),
      timeout: String(testParams.timeout),
    });
  }, [testParams]);

  useEffect(() => {
    if (benchmarkRunning) {
      return;
    }

    const errs = validate(form);
    if (Object.keys(errs).length > 0) {
      return;
    }

    const timer = window.setTimeout(() => {
      const params: TestParams = {
        queries: parseInt(form.queries, 10),
        warmup: parseInt(form.warmup, 10),
        concurrency: parseInt(form.concurrency, 10),
        timeout: parseFloat(form.timeout),
      };

      setTestParams(params);
      SetTestParams(params).catch((err) => {
        setError(getErrorMessage(err, '保存测试参数失败。'));
      });
    }, 300);

    return () => window.clearTimeout(timer);
  }, [benchmarkRunning, form, setError, setTestParams]);

  const validate = (f: typeof form): FieldError => {
    const e: FieldError = {};
    const q = parseInt(f.queries, 10);
    if (isNaN(q) || q <= 0 || !Number.isInteger(q)) e.queries = '须为正整数';
    const w = parseInt(f.warmup, 10);
    if (isNaN(w) || w <= 0 || !Number.isInteger(w)) e.warmup = '须为正整数';
    const c = parseInt(f.concurrency, 10);
    if (isNaN(c) || c <= 0 || !Number.isInteger(c)) e.concurrency = '须为正整数';
    const t = parseFloat(f.timeout);
    if (isNaN(t) || t <= 0) e.timeout = '须为正数';
    return e;
  };

  const handleChange = (field: keyof typeof form, value: string) => {
    const next = { ...form, [field]: value };
    setForm(next);
    setErrors(validate(next));
  };

  const fields: { key: keyof typeof form; label: string; unit: string }[] = [
    { key: 'queries', label: '查询次数', unit: '次/域名' },
    { key: 'warmup', label: '预热次数', unit: '次/服务器' },
    { key: 'concurrency', label: '最大并发', unit: '' },
    { key: 'timeout', label: '超时时间', unit: '秒' },
  ];

  return (
    <div className="test-params-form">
      <span className="section-label">测试参数</span>
      <div className="params-grid">
        {fields.map((f) => (
          <div key={f.key} className="param-field">
            <label className="param-label">
              {f.label}
              {f.unit && <span className="param-unit">({f.unit})</span>}
            </label>
            <input
              className={`param-input ${errors[f.key] ? 'has-error' : ''}`}
              type="text"
              value={form[f.key]}
              onChange={(e) => handleChange(f.key, e.target.value)}
              disabled={benchmarkRunning}
            />
            {errors[f.key] && (
              <span className="param-error">{errors[f.key]}</span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
