import { useAppStore } from '../../store/useAppStore';
import './ErrorBanner.css';

export function ErrorBanner() {
  const errorMsg = useAppStore((s) => s.errorMsg);
  const clearError = useAppStore((s) => s.clearError);

  if (!errorMsg) return null;

  return (
    <div className="error-banner" role="alert">
      <span className="error-banner__icon">⚠</span>
      <span className="error-banner__msg">{errorMsg}</span>
      <button className="error-banner__close" onClick={clearError} aria-label="关闭">✕</button>
    </div>
  );
}
