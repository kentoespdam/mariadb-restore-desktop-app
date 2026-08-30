import { useVirtualizer } from '@tanstack/react-virtual';
import { ArrowLeft, CheckCircle2, RotateCcw, XCircle } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { type CatalogObject, cancelRestore, listCatalogObjects, startPartialRestore } from '@/api';
import { Button } from '@/components/Button';
import { ProgressBar } from '@/components/ProgressBar';
import { navigate, useHashRoute } from '@/hooks/useHashRoute';
import { useWailsEvent } from '@/hooks/useWailsEvent';
import { ObjectGridToolbar } from './ObjectGridToolbar';
import { ObjectRow } from './ObjectRow';

type Phase = 'loading' | 'ready' | 'restoring' | 'done' | 'error';

export function ObjectGrid() {
  const { query } = useHashRoute();
  const filePath = query.get('file') ?? '';
  const profileId = query.get('profile') ?? '';

  const [phase, setPhase] = useState<Phase>('loading');
  const [all, setAll] = useState<CatalogObject[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [filter, setFilter] = useState('');
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [includeRoutines, setIncludeRoutines] = useState(true);
  const [includeTriggers, setIncludeTriggers] = useState(true);
  const [includeEvents, setIncludeEvents] = useState(true);
  const [jobId, setJobId] = useState<string | null>(null);
  const [soFar, setSoFar] = useState(0);
  const [total, setTotal] = useState(0);
  const [result, setResult] = useState<string | null>(null);

  const load = async () => {
    setPhase('loading');
    setErr(null);
    try {
      const objs = await listCatalogObjects(filePath);
      setAll(objs);
      setSelected(new Set(objs.map((o) => o.id)));
      setPhase('ready');
    } catch (e) {
      setErr(String(e));
      setPhase('error');
    }
  };

  // biome-ignore lint/correctness/useExhaustiveDependencies: filePath is the only input; load is stable.
  useEffect(() => {
    if (filePath) load();
  }, [filePath]);

  useWailsEvent<{ jobId: string; soFar: number; total: number }>(
    'restore:progress',
    (p) => {
      if (p.jobId !== jobId) return;
      setSoFar(p.soFar);
      setTotal(p.total);
    },
    { throttleMs: 150 },
  );

  useWailsEvent<{ jobId: string; status: 'success' | 'error'; message?: string }>(
    'restore:done',
    (p) => {
      if (p.jobId !== jobId) return;
      if (p.status === 'success') {
        setPhase('done');
        setResult(`Restored ${selected.size} objects`);
      } else {
        setPhase('error');
        setErr(p.message ?? 'Restore failed');
      }
    },
  );

  const visible = useMemo(() => {
    const f = filter.trim().toLowerCase();
    return all.filter((o) => {
      if (o.type === 'ROUTINE' && !includeRoutines) return false;
      if (o.type === 'TRIGGER' && !includeTriggers) return false;
      if (o.type === 'EVENT' && !includeEvents) return false;
      if (!f) return true;
      return o.database.toLowerCase().includes(f) || o.name.toLowerCase().includes(f);
    });
  }, [all, filter, includeRoutines, includeTriggers, includeEvents]);

  const parentRef = useRef<HTMLDivElement>(null);
  const rowVirtualizer = useVirtualizer({
    count: visible.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 32,
    overscan: 10,
  });

  const toggle = (id: number) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const onSelectAll = () => setSelected(new Set(visible.map((o) => o.id)));
  const onDeselectAll = () => setSelected(new Set());

  const onRestore = async () => {
    setErr(null);
    setResult(null);
    setSoFar(0);
    setTotal(0);
    setPhase('restoring');
    try {
      const handle = await startPartialRestore({
        filePath,
        profileId,
        selectedIds: Array.from(selected),
        includeRoutines,
        includeTriggers,
        includeEvents,
      });
      setJobId(handle.jobId);
    } catch (e) {
      setPhase('error');
      setErr(String(e));
    }
  };

  const onCancel = async () => {
    if (!jobId) return;
    try {
      await cancelRestore(jobId);
    } catch (e) {
      setErr(String(e));
    }
  };

  if (!filePath) {
    return (
      <section className="space-y-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-white">Partial Restore</h2>
          <p className="mt-2 text-slate-400 text-sm">
            No dump file selected. Start from the Restore page.
          </p>
        </div>
        <Button onClick={() => navigate('/restore')}>
          <ArrowLeft className="size-4" />
          Back to Restore
        </Button>
      </section>
    );
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-white">Partial Restore</h2>
          <div className="mt-1 flex items-center gap-2 text-xs text-slate-400">
            <span className="font-mono bg-slate-900 border border-slate-800 px-2 py-0.5 rounded text-sky-300 truncate max-w-md">
              {filePath}
            </span>
            <span>•</span>
            <span className="text-slate-300 font-semibold">{all.length} objects</span>
          </div>
        </div>
        <Button variant="ghost" onClick={() => navigate('/restore')}>
          <ArrowLeft className="size-4" />
          Back
        </Button>
      </div>

      {err && (
        <div className="flex items-start gap-3 p-4 bg-rose-500/10 border border-rose-500/20 rounded-2xl text-rose-300 text-sm">
          <XCircle className="size-5 text-rose-400 shrink-0 mt-0.5" />
          <p>{err}</p>
        </div>
      )}

      {phase === 'loading' && (
        <div className="flex items-center gap-3 p-6 bg-slate-900/80 border border-slate-800 rounded-2xl">
          <div className="size-5 border-2 border-sky-400 border-t-transparent rounded-full animate-spin shrink-0" />
          <p className="text-sm text-slate-300 font-medium">Loading catalog…</p>
        </div>
      )}

      {phase !== 'loading' && (
        <div className="space-y-4">
          <ObjectGridToolbar
            filter={filter}
            onFilterChange={setFilter}
            includeRoutines={includeRoutines}
            onToggleRoutines={setIncludeRoutines}
            includeTriggers={includeTriggers}
            onToggleTriggers={setIncludeTriggers}
            includeEvents={includeEvents}
            onToggleEvents={setIncludeEvents}
            onSelectAll={onSelectAll}
            onDeselectAll={onDeselectAll}
          />

          {/* Virtualized Grid Card */}
          <div className="bg-slate-900/90 border border-slate-800 rounded-2xl overflow-hidden shadow-sm">
            <div className="flex items-center gap-3 px-3.5 h-8 bg-slate-850 border-b border-slate-800 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
              <span className="w-4 shrink-0" />
              <span className="size-3.5 shrink-0" />
              <span className="w-36 shrink-0">Database</span>
              <span className="flex-1">Object Name</span>
              <span className="w-20 text-center">Type</span>
              <span className="w-20 text-right">Offset Size</span>
            </div>

            <div ref={parentRef} className="h-[52vh] overflow-auto bg-slate-900/60">
              <div
                style={{
                  height: `${rowVirtualizer.getTotalSize()}px`,
                  position: 'relative',
                }}
              >
                {rowVirtualizer.getVirtualItems().map((vRow) => {
                  const obj = visible[vRow.index];
                  return (
                    <div
                      key={obj.id}
                      style={{
                        position: 'absolute',
                        top: 0,
                        left: 0,
                        width: '100%',
                        transform: `translateY(${vRow.start}px)`,
                      }}
                    >
                      <ObjectRow obj={obj} selected={selected.has(obj.id)} onToggle={toggle} />
                    </div>
                  );
                })}
              </div>
            </div>
          </div>

          {/* Action Bar */}
          <div className="flex flex-wrap items-center justify-between gap-4 p-4 bg-slate-900/80 border border-slate-800 rounded-2xl shadow-sm">
            <div className="flex items-center gap-3">
              <Button
                onClick={onRestore}
                disabled={phase === 'restoring' || selected.size === 0}
                className="bg-sky-600 hover:bg-sky-500 active:bg-sky-700 shadow-sky-950/50"
              >
                <RotateCcw className="size-4" />
                Restore selected ({selected.size})
              </Button>
              <Button variant="outline" size="md" onClick={load}>
                Re-analyze
              </Button>
              {phase === 'restoring' && (
                <Button variant="ghost" size="md" onClick={onCancel}>
                  Cancel
                </Button>
              )}
            </div>

            <div className="text-xs text-slate-400">
              Showing <strong className="text-slate-200">{visible.length}</strong> of{' '}
              <strong className="text-slate-200">{all.length}</strong> objects
            </div>
          </div>

          {phase === 'restoring' && (
            <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-6 shadow-sm">
              <ProgressBar soFar={soFar} total={total} label="Restore progress" />
            </div>
          )}

          {phase === 'done' && result && (
            <div className="flex items-start gap-3 p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-2xl text-emerald-300 text-sm">
              <CheckCircle2 className="size-5 text-emerald-400 shrink-0 mt-0.5" />
              <p className="font-semibold">{result}</p>
            </div>
          )}
        </div>
      )}
    </section>
  );
}
