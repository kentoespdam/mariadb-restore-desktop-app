// ponytail: minimal determinate progress bar. total=0 renders the bar
// in indeterminate "scanning" mode (a slow pulse via Tailwind animate).
// No external lib; raw utilities only per ADR-0008.
export function ProgressBar({
  soFar,
  total,
  label,
}: {
  soFar: number;
  total: number;
  label?: string;
}) {
  const pct = total > 0 ? Math.min(100, Math.round((soFar / total) * 100)) : 0;
  const indeterminate = total <= 0;

  return (
    <div className="w-full">
      {label && (
        <div className="flex justify-between text-xs text-slate-400 mb-1">
          <span>{label}</span>
          <span>{indeterminate ? '…' : `${pct}%`}</span>
        </div>
      )}
      <div
        className="h-2 bg-slate-700 rounded overflow-hidden"
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={total > 0 ? total : 100}
        aria-valuenow={indeterminate ? undefined : soFar}
      >
        {indeterminate ? (
          <div className="h-full w-1/3 bg-slate-400 animate-pulse" />
        ) : (
          <div
            className="h-full bg-slate-300 transition-[width] duration-150"
            style={{ width: `${pct}%` }}
          />
        )}
      </div>
    </div>
  );
}
