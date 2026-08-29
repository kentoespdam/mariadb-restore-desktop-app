// Restore workflow: FE contract.
//
// Full Restore: file -> mariadb CLI, no scan (CONTEXT: "triggered
// when the user clicks Restore immediately after selecting a file,
// bypassing the Analyze step").
// Partial Restore: file -> Analyze -> catalog -> user picks objects
//                 -> Virtual Streamer pipes selected byte ranges.

export type ObjectType = 'CREATE_TABLE' | 'INSERT' | 'USE' | 'ROUTINE' | 'TRIGGER' | 'EVENT';

export interface CatalogObject {
  id: number;
  database: string;
  name: string;
  type: ObjectType;
  startByte: number;
  endByte: number;
}

export interface FullRestoreRequest {
  filePath: string;
  profileId: string;
}

export interface PartialRestoreRequest {
  filePath: string;
  profileId: string;
  selectedIds: number[];
  includeRoutines: boolean;
  includeTriggers: boolean;
  includeEvents: boolean;
}

export interface RestoreHandle {
  jobId: string;
  cancel: () => Promise<void>;
}

function notImplemented(name: string): never {
  throw new Error(`not implemented: ${name}`);
}

export async function startFullRestore(_req: FullRestoreRequest): Promise<RestoreHandle> {
  notImplemented('startFullRestore');
}

export async function analyzeDump(_filePath: string): Promise<{ objectCount: number }> {
  notImplemented('analyzeDump');
}

export async function listCatalogObjects(_filePath: string): Promise<CatalogObject[]> {
  notImplemented('listCatalogObjects');
}

export async function startPartialRestore(_req: PartialRestoreRequest): Promise<RestoreHandle> {
  notImplemented('startPartialRestore');
}

export async function cancelRestore(_jobId: string): Promise<void> {
  notImplemented('cancelRestore');
}
