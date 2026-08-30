import { Search } from 'lucide-react';
import { Button } from '@/components/Button';

interface ObjectGridToolbarProps {
  filter: string;
  onFilterChange: (val: string) => void;
  includeRoutines: boolean;
  onToggleRoutines: (val: boolean) => void;
  includeTriggers: boolean;
  onToggleTriggers: (val: boolean) => void;
  includeEvents: boolean;
  onToggleEvents: (val: boolean) => void;
  onSelectAll: () => void;
  onDeselectAll: () => void;
}

const inputClass =
  'bg-slate-900/90 border-slate-700/80 hover:border-slate-600 focus:border-sky-500 h-9 rounded-xl border px-3 py-1.5 text-xs text-white w-full transition-all focus:outline-none focus:ring-2 focus:ring-sky-500/30 placeholder:text-slate-500';

export function ObjectGridToolbar({
  filter,
  onFilterChange,
  includeRoutines,
  onToggleRoutines,
  includeTriggers,
  onToggleTriggers,
  includeEvents,
  onToggleEvents,
  onSelectAll,
  onDeselectAll,
}: ObjectGridToolbarProps) {
  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-4 shadow-sm flex flex-wrap items-center gap-3">
      <div className="relative flex-1 min-w-56">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-3.5 text-slate-400" />
        <input
          aria-label="Filter"
          value={filter}
          onChange={(e) => onFilterChange(e.target.value)}
          placeholder="Filter by database or name…"
          className={`${inputClass} pl-9`}
        />
      </div>

      <div className="flex items-center gap-4 text-xs text-slate-300 border-x border-slate-800 px-3">
        <label className="flex items-center gap-1.5 cursor-pointer">
          <input
            type="checkbox"
            checked={includeRoutines}
            onChange={(e) => onToggleRoutines(e.target.checked)}
            className="rounded border-slate-700 bg-slate-900 text-sky-500"
          />
          Routines
        </label>
        <label className="flex items-center gap-1.5 cursor-pointer">
          <input
            type="checkbox"
            checked={includeTriggers}
            onChange={(e) => onToggleTriggers(e.target.checked)}
            className="rounded border-slate-700 bg-slate-900 text-sky-500"
          />
          Triggers
        </label>
        <label className="flex items-center gap-1.5 cursor-pointer">
          <input
            type="checkbox"
            checked={includeEvents}
            onChange={(e) => onToggleEvents(e.target.checked)}
            className="rounded border-slate-700 bg-slate-900 text-sky-500"
          />
          Events
        </label>
      </div>

      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={onSelectAll}>
          Select all
        </Button>
        <Button variant="ghost" size="sm" onClick={onDeselectAll}>
          Deselect all
        </Button>
      </div>
    </div>
  );
}
