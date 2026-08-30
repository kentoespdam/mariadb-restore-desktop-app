// ponytail: uses Wails native file dialog to get the full filesystem
// path directly, avoiding the browser's <input type="file"> which
// only exposes the filename for security reasons.
import { OpenDumpFileDialog } from '../../wailsjs/go/app/App';

export interface FilePickerProps {
  id: string;
  accept?: string;
  onChange: (path: string) => void;
}

export function FilePicker({ id, onChange }: FilePickerProps) {
  const handleClick = async () => {
    const path = await OpenDumpFileDialog();
    if (path) onChange(path);
  };

  return (
    <button
      id={id}
      type="button"
      onClick={handleClick}
      className="block w-full text-sm text-slate-300 bg-slate-900 border border-slate-700 rounded-md px-3 py-1.5 text-left hover:bg-slate-800"
    >
      Choose file...
    </button>
  );
}
