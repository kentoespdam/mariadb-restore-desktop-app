// Backup workflow: real Wails bindings. The Go side runs
// mariadb-dump in a subprocess and streams backup:progress +
// backup:done events.
import { CancelBackup, StartBackup } from '../../wailsjs/go/app/App';
import type { app } from '../../wailsjs/go/models';

export type BackupRequest = app.BackupRequest;
export type BackupHandle = { jobId: string; cancel: () => Promise<void> };

export async function startBackup(req: BackupRequest): Promise<BackupHandle> {
  const jobId = await StartBackup(req);
  return {
    jobId,
    cancel: async () => {
      await CancelBackup(jobId);
    },
  };
}

export async function cancelBackup(jobID: string): Promise<void> {
  await CancelBackup(jobID);
}
