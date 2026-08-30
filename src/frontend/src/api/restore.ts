// Restore workflow: real Wails bindings. The Go side runs
// mariadb CLI in a subprocess and streams restore:progress +
// restore:done events.
//
// Full Restore: file -> mariadb CLI, no scan (CONTEXT).
// Partial Restore: file -> Analyze -> catalog -> user picks
//                 objects -> Virtual Streamer pipes selected
//                 byte ranges.
import {
  AnalyzeDump,
  CancelRestore,
  ListCatalogObjects,
  StartFullRestore,
  StartPartialRestore,
} from '../../wailsjs/go/app/App';
import type { app } from '../../wailsjs/go/models';

export type ObjectType = 'CREATE_TABLE' | 'INSERT' | 'USE' | 'ROUTINE' | 'TRIGGER' | 'EVENT';

export type CatalogObject = app.CatalogObject;
export type FullRestoreRequest = app.RestoreRequest;
export type PartialRestoreRequest = app.PartialRestoreRequest;
export type RestoreHandle = { jobId: string; cancel: () => Promise<void> };

export async function startFullRestore(req: FullRestoreRequest): Promise<RestoreHandle> {
  const jobId = await StartFullRestore(req);
  return {
    jobId,
    cancel: async () => {
      await CancelRestore(jobId);
    },
  };
}

export async function analyzeDump(filePath: string): Promise<{ objectCount: number }> {
  const count = await AnalyzeDump(filePath);
  return { objectCount: count };
}

export async function listCatalogObjects(filePath: string): Promise<CatalogObject[]> {
  return ListCatalogObjects(filePath);
}

export async function startPartialRestore(req: PartialRestoreRequest): Promise<RestoreHandle> {
  const jobId = await StartPartialRestore(req);
  return {
    jobId,
    cancel: async () => {
      await CancelRestore(jobId);
    },
  };
}

export async function cancelRestore(jobID: string): Promise<void> {
  await CancelRestore(jobID);
}
