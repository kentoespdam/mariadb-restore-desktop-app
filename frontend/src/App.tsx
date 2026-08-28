import {useEffect, useState} from 'react';
import {Init, InitFresh, ResetAndInit} from '../wailsjs/go/main/App';
import {Quit} from '../wailsjs/runtime/runtime';
import {main} from '../wailsjs/go/models';
import './App.css';
import './servers.css';
import {Sidebar} from './components/Sidebar';
import {SmartRecoveryModal} from './components/SmartRecoveryModal';
import {ServersPanel} from './panels/Servers';
import {BackupPanel} from './panels/Backup';
import {RestorePanel} from './panels/Restore';
import {SettingsPanel} from './panels/Settings';

export type Panel = 'servers' | 'backup' | 'restore' | 'settings';

function App() {
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [showRecovery, setShowRecovery] = useState(false);
    const [activePanel, setActivePanel] = useState<Panel>('servers');

    const startApp = async () => {
        try {
            setLoading(true);
            setError(null);
            const result = await Init();
            switch (result.status) {
                case 'ready':
                    setLoading(false);
                    break;
                case 'fresh':
                    await InitFresh();
                    setLoading(false);
                    break;
                case 'needs_recovery':
                    setShowRecovery(true);
                    setLoading(false);
                    break;
                default:
                    setLoading(false);
                    break;
            }
        } catch (err: unknown) {
            console.error('Initialization failed:', err);
            setError(String(err));
            setLoading(false);
        }
    };

    useEffect(() => {
        startApp();
    }, []);

    function handleCancel() {
        Quit();
    }

    async function handleReset() {
        await ResetAndInit();
        setShowRecovery(false);
    }

    if (loading) {
        return <div className="loading">Initializing…</div>;
    }

    if (error) {
        return (
            <div className="loading" style={{flexDirection: 'column', gap: '1rem'}}>
                <div style={{fontSize: '1.2rem', color: '#e94560'}}>Initialization Error</div>
                <div style={{color: '#aaa', maxWidth: '400px', textAlign: 'center'}}>{error}</div>
                <button className="btn btn-primary" onClick={startApp}>Retry</button>
            </div>
        );
    }

    return (
        <div id="App">
            <Sidebar active={activePanel} onChange={setActivePanel}/>
            <main className="content">
                {activePanel === 'servers' && <ServersPanel/>}
                {activePanel === 'backup' && <BackupPanel/>}
                {activePanel === 'restore' && <RestorePanel/>}
                {activePanel === 'settings' && <SettingsPanel/>}
            </main>
            {showRecovery && (
                <SmartRecoveryModal onCancel={handleCancel} onReset={handleReset}/>
            )}
        </div>
    );
}

export default App;
