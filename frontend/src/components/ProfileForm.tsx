import {useState} from 'react';
import {profile} from '../../wailsjs/go/models';
import {TestConnection, SaveProfile, DeleteProfile} from '../../wailsjs/go/main/App';

interface ProfileFormProps {
    profile: profile.Profile | null;
    onSaved: () => void;
    onDeleted: () => void;
}

const SSL_MODES = ['disabled', 'required', 'verify-ca', 'verify-full'];

export function ProfileForm({profile: p, onSaved, onDeleted}: ProfileFormProps) {
    const [form, setForm] = useState<profile.Profile>(() => p ?? new profile.Profile());
    const [testResult, setTestResult] = useState<profile.TestResult | null>(null);
    const [testing, setTesting] = useState(false);
    const [saving, setSaving] = useState(false);

    // Sync when selected profile changes
    const prevId = useState(p?.id);
    if (p && p.id !== prevId[0]) {
        prevId[0] = p.id;
        setForm({...p});
        setTestResult(null);
    }

    function update(field: keyof profile.Profile, value: string | number) {
        setForm((f) => ({...f, [field]: value}));
        setTestResult(null);
    }

    async function handleTest() {
        setTesting(true);
        setTestResult(null);
        try {
            const result = await TestConnection(form);
            setTestResult(result);
        } catch (e: unknown) {
            setTestResult({ok: false, error: String(e)});
        } finally {
            setTesting(false);
        }
    }

    async function handleSave() {
        setSaving(true);
        try {
            await SaveProfile(form);
            onSaved();
        } catch (e: unknown) {
            alert(`Save failed: ${e}`);
        } finally {
            setSaving(false);
        }
    }

    async function handleDelete() {
        if (!form.id || !confirm('Delete this profile?')) return;
        try {
            await DeleteProfile(form.id);
            onDeleted();
        } catch (e: unknown) {
            alert(`Delete failed: ${e}`);
        }
    }

    if (!p) {
        return (
            <div className="profile-form-empty">
                <p>Select a profile or create a new one.</p>
            </div>
        );
    }

    return (
        <div className="profile-form">
            <div className="form-row">
                <label>Name</label>
                <input value={form.name} onChange={(e) => update('name', e.target.value)}/>
            </div>
            <div className="form-row form-row-inline">
                <div className="form-field">
                    <label>Host</label>
                    <input value={form.host} onChange={(e) => update('host', e.target.value)}/>
                </div>
                <div className="form-field form-field-small">
                    <label>Port</label>
                    <input type="number" value={form.port} onChange={(e) => update('port', Number(e.target.value))}/>
                </div>
            </div>
            <div className="form-row form-row-inline">
                <div className="form-field">
                    <label>Username</label>
                    <input value={form.username} onChange={(e) => update('username', e.target.value)}/>
                </div>
                <div className="form-field">
                    <label>Password</label>
                    <input type="password" placeholder={form.id ? '•••••• (leave empty to keep)' : ''}
                           onChange={(e) => update('password', e.target.value)}/>
                </div>
            </div>

            <fieldset className="ssl-section">
                <legend>SSL</legend>
                <div className="form-row">
                    <label>SSL Mode</label>
                    <select value={form.sslMode} onChange={(e) => update('sslMode', e.target.value)}>
                        {SSL_MODES.map((m) => <option key={m} value={m}>{m}</option>)}
                    </select>
                </div>
                {form.sslMode !== 'disabled' && (
                    <>
                        <div className="form-row">
                            <label>CA Certificate</label>
                            <input value={form.sslCa} placeholder="/path/to/ca.pem"
                                   onChange={(e) => update('sslCa', e.target.value)}/>
                        </div>
                        <div className="form-row">
                            <label>Client Certificate</label>
                            <input value={form.sslCert} placeholder="/path/to/client-cert.pem"
                                   onChange={(e) => update('sslCert', e.target.value)}/>
                        </div>
                        <div className="form-row">
                            <label>Client Key</label>
                            <input value={form.sslKey} placeholder="/path/to/client-key.pem"
                                   onChange={(e) => update('sslKey', e.target.value)}/>
                        </div>
                    </>
                )}
            </fieldset>

            <div className="form-actions">
                <button className="btn btn-secondary" onClick={handleTest} disabled={testing}>
                    {testing ? 'Testing…' : '🔌 Test Connection'}
                </button>
                <button className="btn btn-primary" onClick={handleSave} disabled={saving}>
                    {saving ? 'Saving…' : '💾 Save'}
                </button>
                {form.id > 0 && (
                    <button className="btn btn-danger" onClick={handleDelete}>🗑️ Delete</button>
                )}
            </div>

            {testResult && (
                <div className={`test-result ${testResult.ok ? 'success' : 'error'}`}>
                    {testResult.ok
                        ? `✅ Connected — MariaDB ${testResult.version}`
                        : `❌ ${testResult.error}`}
                </div>
            )}
        </div>
    );
}
