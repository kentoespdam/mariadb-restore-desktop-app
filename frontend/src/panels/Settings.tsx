import {useCallback, useEffect, useState} from 'react';
import {settings} from '../../wailsjs/go/models';
import {GetSettings, SetMariaDBPath, SetMariaDBDumpPath, DiscoverBinaries} from '../../wailsjs/go/main/App';
import '../settings.css';

export function SettingsPanel() {
    const [s, setS] = useState<settings.Settings | null>(null);
    const [mariadbInput, setMariadbInput] = useState('');
    const [dumpInput, setDumpInput] = useState('');
    const [saving, setSaving] = useState(false);
    const [msg, setMsg] = useState('');

    const load = useCallback(async () => {
        try {
            const settings_ = await GetSettings();
            setS(settings_);
            setMariadbInput(settings_.mariadbPath);
            setDumpInput(settings_.mariadbdumpPath);
        } catch (e) {
            console.error('Failed to load settings:', e);
        }
    }, []);

    useEffect(() => { load(); }, [load]);

    async function handleSaveMariaDB() {
        setSaving(true);
        setMsg('');
        try {
            await SetMariaDBPath(mariadbInput);
            setMsg('✅ mariadb path saved');
            load();
        } catch (e: unknown) {
            setMsg(`❌ ${e}`);
        } finally {
            setSaving(false);
        }
    }

    async function handleSaveDump() {
        setSaving(true);
        setMsg('');
        try {
            await SetMariaDBDumpPath(dumpInput);
            setMsg('✅ mariadb-dump path saved');
            load();
        } catch (e: unknown) {
            setMsg(`❌ ${e}`);
        } finally {
            setSaving(false);
        }
    }

    async function handleRediscover() {
        const discovered = await DiscoverBinaries();
        setS(discovered);
        setMariadbInput(discovered.mariadbPath);
        setDumpInput(discovered.mariadbdumpPath);
        setMsg('🔍 Re-discovery complete');
    }

    if (!s) return <div className="loading">Loading settings…</div>;

    return (
        <div className="panel settings-panel">
            <h1>⚙️ Settings</h1>

            {!s.mariadbFound && (
                <div className="settings-banner">
                    ⚠️ mariadb CLI not found — configure path below
                </div>
            )}

            <section className="settings-section">
                <h2>MariaDB Binary Paths</h2>
                <p className="settings-hint">
                    Paths are auto-discovered. Override below if needed.
                </p>

                <div className="settings-row">
                    <label>mariadb executable</label>
                    <div className="settings-input-row">
                        <input
                            value={mariadbInput}
                            onChange={(e) => setMariadbInput(e.target.value)}
                            placeholder="/usr/bin/mariadb"
                        />
                        <button className="btn btn-primary btn-sm" onClick={handleSaveMariaDB} disabled={saving}>
                            Save
                        </button>
                    </div>
                    {s.mariadbPath && (
                        <div className="settings-resolved">Found: {s.mariadbPath}</div>
                    )}
                </div>

                <div className="settings-row">
                    <label>mariadb-dump executable</label>
                    <div className="settings-input-row">
                        <input
                            value={dumpInput}
                            onChange={(e) => setDumpInput(e.target.value)}
                            placeholder="/usr/bin/mariadb-dump"
                        />
                        <button className="btn btn-primary btn-sm" onClick={handleSaveDump} disabled={saving}>
                            Save
                        </button>
                    </div>
                    {s.mariadbdumpPath && (
                        <div className="settings-resolved">Found: {s.mariadbdumpPath}</div>
                    )}
                </div>

                <button className="btn btn-secondary" onClick={handleRediscover}>
                    🔍 Re-discover
                </button>

                {msg && <div className="settings-msg">{msg}</div>}
            </section>

            <section className="settings-section">
                <h2>About</h2>
                <p>
                    Portable MariaDB Backup &amp; Restore Tool
                </p>
                <p className="settings-license">
                    MariaDB binaries are licensed under GPL/LGPL.
                    See <a href="https://mariadb.com/kb/en/library/licensing-faq/" target="_blank" rel="noreferrer">
                    mariadb.com/kb/en/licensing-faq</a> for details.
                </p>
            </section>
        </div>
    );
}
