import {profile} from '../../wailsjs/go/models';

interface ProfileListProps {
    profiles: profile.Profile[];
    selectedId: number | null;
    onSelect: (id: number) => void;
    onNew: () => void;
}

export function ProfileList({profiles, selectedId, onSelect, onNew}: ProfileListProps) {
    return (
        <div className="profile-list">
            <div className="profile-list-header">
                <span>Profiles</span>
                <button className="btn btn-sm" onClick={onNew}>+ New</button>
            </div>
            <div className="profile-list-items">
                {profiles.length === 0 && (
                    <div className="profile-list-empty">No profiles yet</div>
                )}
                {profiles.map((p) => (
                    <button
                        key={p.id}
                        className={`profile-list-item ${p.id === selectedId ? 'selected' : ''}`}
                        onClick={() => onSelect(p.id)}
                    >
                        <span className="profile-name">{p.name}</span>
                        <span className="profile-host">{p.host}:{p.port}</span>
                    </button>
                ))}
            </div>
        </div>
    );
}
