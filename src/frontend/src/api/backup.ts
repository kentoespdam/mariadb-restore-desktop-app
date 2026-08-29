// Backup workflow: FE contract. The real Go binding
// (src/backend/features/backup/) lands in a follow-up plan; until then
// stubs surface the missing implementation in dev.
//
// ponytail: stub-throws-on-call so a forgotten real implementation
// cannot pass silently. Re-throw the same string the FE expects to
// see in production once a real BE slice wires the same name.

export interface BackupRequest {
  profileId: string;
  databases: string[];
  outputPath: string;
}

export interface BackupHandle {
  jobId: string;
  cancel: () => Promise<void>;
}

function notImplemented(name: string): never {
  throw new Error(`not implemented: ${name}`);
}

export async function startBackup(_req: BackupRequest): Promise<BackupHandle> {
  notImplemented('startBackup');
}

export async function cancelBackup(_jobId: string): Promise<void> {
  notImplemented('cancelBackup');
}
