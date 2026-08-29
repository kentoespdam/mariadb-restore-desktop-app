// Thin re-export of the wails-generated profile bindings.
// The shape is owned by src/backend/features/profile (Go).
export {
  CreateServerProfile,
  DeleteServerProfile,
  ListServerProfiles,
  UpdateServerProfile,
} from '../../wailsjs/go/app/App';
export type { profile } from '../../wailsjs/go/models';
