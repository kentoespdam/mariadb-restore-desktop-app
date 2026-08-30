import { FolderOpen } from 'lucide-react';
import { OpenDumpFileDialog } from '../../wailsjs/go/app/App';

export interface FilePickerProps {
  id: string;
  accept?: string;
  onChange: (path: string) => void;
  disabled?: boolean;
}

export function FilePicker({ id, onChange, disabled }: FilePickerProps) {
  const handleClick = async () => {
    const path = await OpenDumpFileDialog();
    if (path) onChange(path);
  };

  return (
    <button
      id={id}
      type="button"
      onClick={handleClick}
      disabled={disabled}
      className="group flex items-center justify-between w-full text-sm text-slate-300 bg-slate-900/90 border border-slate-700/80 hover:border-sky-500/50 rounded-lg px-3.5 py-2 text-left hover:bg-slate-850 transition-all duration-150 focus:outline-none focus:ring-2 focus:ring-sky-500/40 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
    >
      <span className="flex items-center gap-2 text-slate-400 group-hover:text-slate-200">
        <FolderOpen className="size-4 text-sky-400 shrink-0 group-hover:scale-110 transition-transform" />
        <span>Choose file...</span>
      </span>
      <span className="text-xs bg-slate-800 text-slate-400 px-2 py-0.5 rounded border border-slate-700 font-mono">
        Browse
      </span>
    </button>
  );
}
