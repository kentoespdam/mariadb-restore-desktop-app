function fmtBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

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
    <div className="w-full bg-slate-900/80 border border-slate-800 rounded-xl p-3.5 shadow-sm">
      <div className="flex justify-between items-center text-xs text-slate-300 font-medium mb-2">
        <span className="flex items-center gap-1.5">
          <span className="inline-block size-2 rounded-full bg-sky-400 animate-ping" />
          {label ?? 'Processing…'}
        </span>
        <div className="flex items-center gap-2 tabular-nums">
          {!indeterminate && total > 1024 && (
            <span className="text-slate-400 font-normal">
              {fmtBytes(soFar)} / {fmtBytes(total)}
            </span>
          )}
          <span className="font-semibold text-sky-400">
            {indeterminate ? (soFar > 0 ? fmtBytes(soFar) : '…') : `${pct}%`}
          </span>
        </div>
      </div>
      <div
        className="h-2.5 bg-slate-800 rounded-full overflow-hidden relative"
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={total > 0 ? total : 100}
        aria-valuenow={indeterminate ? undefined : soFar}
      >
        {indeterminate ? (
          <div className="h-full w-1/3 bg-gradient-to-r from-sky-500 via-indigo-400 to-sky-500 rounded-full animate-pulse" />
        ) : (
          <div
            className="h-full bg-gradient-to-r from-sky-500 to-emerald-400 rounded-full transition-all duration-200 shadow-sm shadow-sky-500/50"
            style={{ width: `${pct}%` }}
          />
        )}
      </div>
    </div>
  );
}
