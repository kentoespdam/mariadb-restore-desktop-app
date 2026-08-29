// ponytail: thin wrapper around <input type="file"> that reports the
// picked filename up via onChange. We don't route through
// runtime.OpenFileDialog yet because the FE plan is standalone and
// the BE-side file-resolution story is out of scope.
import type { ChangeEvent } from 'react';

export interface FilePickerProps {
  id: string;
  accept?: string;
  onChange: (path: string) => void;
}

export function FilePicker({ id, accept, onChange }: FilePickerProps) {
  return (
    <input
      id={id}
      type="file"
      accept={accept}
      onChange={(e: ChangeEvent<HTMLInputElement>) => {
        const f = e.target.files?.[0];
        if (f) onChange(f.name);
      }}
      className="block w-full text-sm text-slate-300 file:mr-3 file:rounded-md file:border-0 file:bg-slate-700 file:px-3 file:py-1.5 file:text-sm file:text-white hover:file:bg-slate-600"
    />
  );
}
