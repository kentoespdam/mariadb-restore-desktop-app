// Barrel re-export. Screens import from '@/api', never from
// wailsjs/go/app/App directly — this is what lets the missing
// backup/restore/settings bindings be replaced later without
// changing screen code.

export * from './backup';
export * from './profile';
export * from './recovery';
export * from './restore';
export * from './settings';
