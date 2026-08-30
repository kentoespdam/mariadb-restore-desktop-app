import { Database } from 'lucide-react';
import type { CatalogObject } from '@/api';

const TYPE_BADGE: Record<CatalogObject['type'], string> = {
  CREATE_TABLE: 'bg-sky-500/15 text-sky-300 border-sky-500/30',
  INSERT: 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30',
  USE: 'bg-slate-800 text-slate-300 border-slate-700',
  ROUTINE: 'bg-amber-500/15 text-amber-300 border-amber-500/30',
  TRIGGER: 'bg-purple-500/15 text-purple-300 border-purple-500/30',
  EVENT: 'bg-cyan-500/15 text-cyan-300 border-cyan-500/30',
};

function fmtSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

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
    <div
      className={`flex items-center gap-3 px-3.5 h-8 border-b border-slate-800/80 text-xs transition-colors duration-100 ${
        selected ? 'bg-sky-500/[0.04]' : 'hover:bg-slate-800/40'
      }`}
    >
      <input
        type="checkbox"
        checked={selected}
        onChange={() => onToggle(obj.id)}
        aria-label={`Select ${obj.database}.${obj.name}`}
        className="size-3.5 rounded border-slate-700 bg-slate-900 text-sky-500 focus:ring-sky-500/30 cursor-pointer"
      />
      <Database className="size-3.5 text-slate-500 shrink-0" />
      <span className="text-slate-400 font-mono truncate w-36 shrink-0">{obj.database}</span>
      <span className="font-mono text-slate-200 truncate flex-1 font-medium">{obj.name}</span>
      <span
        className={`text-[9px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full border ${TYPE_BADGE[obj.type]}`}
      >
        {obj.type}
      </span>
      <span className="text-slate-400 w-20 text-right tabular-nums font-mono text-[11px]">
        {fmtSize(obj.endByte - obj.startByte)}
      </span>
    </div>
  );
}
