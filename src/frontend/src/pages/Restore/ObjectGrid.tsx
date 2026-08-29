import { useVirtualizer } from '@tanstack/react-virtual';
import { Search } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { type CatalogObject, cancelRestore, listCatalogObjects, startPartialRestore } from '@/api';
import { Button } from '@/components/Button';
import { ProgressBar } from '@/components/ProgressBar';
import { navigate, useHashRoute } from '@/hooks/useHashRoute';
import { useWailsEvent } from '@/hooks/useWailsEvent';
import { ObjectRow } from './ObjectRow';

const inputClass =
  'bg-slate-900 border-slate-700 h-9 rounded-md border px-3 py-1 text-sm text-white w-full';

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
      <section>
        <h2 className="text-xl font-semibold">Partial Restore</h2>
        <p className="mt-2 text-slate-400">No dump file selected. Start from the Restore page.</p>
        <Button className="mt-4" onClick={() => navigate('/restore')}>
          Back to Restore
        </Button>
      </section>
    );
  }

  return (
    <section>
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Partial Restore</h2>
        <Button variant="ghost" onClick={() => navigate('/restore')}>
          Back
        </Button>
      </div>
      <p className="mt-1 text-slate-400 text-sm">
        <span className="font-mono">{filePath}</span> · {all.length} objects
      </p>
      {err && <p className="mt-3 text-red-400 text-sm">{err}</p>}

      {phase === 'loading' && <p className="mt-6 text-slate-400">Loading catalog…</p>}

      {phase !== 'loading' && (
        <>
          <div className="mt-6 flex flex-wrap items-center gap-3 max-w-4xl">
            <div className="relative flex-1 min-w-64">
              <Search className="absolute left-2 top-1/2 -translate-y-1/2 size-4 text-slate-500" />
              <input
                aria-label="Filter"
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
                placeholder="Filter by database or name…"
                className={`${inputClass} pl-8`}
              />
            </div>
            <label className="flex items-center gap-1.5 text-sm text-slate-300">
              <input
                type="checkbox"
                checked={includeRoutines}
                onChange={(e) => setIncludeRoutines(e.target.checked)}
              />
              Routines
            </label>
            <label className="flex items-center gap-1.5 text-sm text-slate-300">
              <input
                type="checkbox"
                checked={includeTriggers}
                onChange={(e) => setIncludeTriggers(e.target.checked)}
              />
              Triggers
            </label>
            <label className="flex items-center gap-1.5 text-sm text-slate-300">
              <input
                type="checkbox"
                checked={includeEvents}
                onChange={(e) => setIncludeEvents(e.target.checked)}
              />
              Events
            </label>
            <Button variant="ghost" onClick={onSelectAll}>
              Select all
            </Button>
            <Button variant="ghost" onClick={onDeselectAll}>
              Deselect all
            </Button>
          </div>

          <div
            ref={parentRef}
            className="mt-4 h-[60vh] max-w-4xl overflow-auto bg-slate-800 border border-slate-700 rounded-lg"
          >
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

          <div className="mt-4 flex gap-2 items-center">
            <Button onClick={onRestore} disabled={phase === 'restoring' || selected.size === 0}>
              Restore selected ({selected.size})
            </Button>
            <Button variant="ghost" onClick={load}>
              Re-analyze
            </Button>
            {phase === 'restoring' && (
              <Button variant="ghost" onClick={onCancel}>
                Cancel
              </Button>
            )}
          </div>

          {phase === 'restoring' && (
            <div className="mt-4 max-w-2xl">
              <ProgressBar soFar={soFar} total={total} label="Restore progress" />
            </div>
          )}
          {phase === 'done' && result && <p className="mt-4 text-emerald-400 text-sm">{result}</p>}
        </>
      )}
    </section>
  );
}
