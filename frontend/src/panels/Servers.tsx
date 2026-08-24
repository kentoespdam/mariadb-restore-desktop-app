import {useCallback, useEffect, useState} from 'react';
import {profile} from '../../wailsjs/go/models';
import {ListProfiles} from '../../wailsjs/go/main/App';
import {ProfileList} from '../components/ProfileList';
import {ProfileForm} from '../components/ProfileForm';

export function ServersPanel() {
    const [profiles, setProfiles] = useState<profile.Profile[]>([]);
    const [selectedId, setSelectedId] = useState<number | null>(null);
    const [creating, setCreating] = useState(false);

    const load = useCallback(async () => {
        try {
            const list = await ListProfiles();
            setProfiles(list);
        } catch (e) {
            console.error('Failed to load profiles:', e);
        }
    }, []);

    useEffect(() => { load(); }, [load]);

    function getSelected(): profile.Profile | null {
        if (creating) {
            const fresh = new profile.Profile();
            fresh.name = 'New Server';
            fresh.host = 'localhost';
            fresh.port = 3306;
            fresh.username = 'root';
            fresh.sslMode = 'disabled';
            return fresh;
        }
        return profiles.find((p) => p.id === selectedId) ?? null;
    }

    function handleNew() {
        setCreating(true);
        setSelectedId(null);
    }

    function handleSaved() {
        setCreating(false);
        load();
        setSelectedId(null);
    }

    function handleDeleted() {
        load();
        setSelectedId(null);
    }

    return (
        <div className="panel servers-panel">
            <ProfileList
                profiles={profiles}
                selectedId={selectedId}
                onSelect={(id) => { setSelectedId(id); setCreating(false); }}
                onNew={handleNew}
            />
            <ProfileForm
                profile={getSelected()}
                onSaved={handleSaved}
                onDeleted={handleDeleted}
            />
        </div>
    );
}
