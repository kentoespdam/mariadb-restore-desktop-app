import { Database } from 'lucide-react';
import type { CatalogObject } from '@/api';

const TYPE_BADGE: Record<CatalogObject['type'], string> = {
  CREATE_TABLE: 'bg-slate-700 text-slate-200',
  INSERT: 'bg-emerald-900 text-emerald-200',
  USE: 'bg-slate-600 text-slate-300',
  ROUTINE: 'bg-amber-900 text-amber-200',
  TRIGGER: 'bg-purple-900 text-purple-200',
  EVENT: 'bg-sky-900 text-sky-200',
};

function fmtSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

// ponytail: pure presentational row, kept under 50 LOC. The grid owns
// selection state and renders these as virtualizer children.
export function ObjectRow({
  obj,
  selected,
  onToggle,
}: {
  obj: CatalogObject;
  selected: boolean;
  onToggle: (id: number) => void;
}) {
  return (
    <div className="flex items-center gap-3 px-3 h-8 border-b border-slate-800 text-sm">
      <input
        type="checkbox"
        checked={selected}
        onChange={() => onToggle(obj.id)}
        aria-label={`Select ${obj.database}.${obj.name}`}
      />
      <Database className="size-3.5 text-slate-500 shrink-0" />
      <span className="text-slate-400 truncate w-40">{obj.database}</span>
      <span className="font-mono text-slate-200 truncate flex-1">{obj.name}</span>
      <span
        className={`text-[10px] uppercase tracking-wide px-1.5 py-0.5 rounded ${TYPE_BADGE[obj.type]}`}
      >
        {obj.type}
      </span>
      <span className="text-slate-500 w-20 text-right tabular-nums">
        {fmtSize(obj.endByte - obj.startByte)}
      </span>
    </div>
  );
}
