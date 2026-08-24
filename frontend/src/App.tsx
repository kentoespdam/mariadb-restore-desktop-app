import {useEffect, useState} from 'react';
import {Init, InitFresh, Recover, ResetAndInit} from '../wailsjs/go/main/App';
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
    const [showRecovery, setShowRecovery] = useState(false);
    const [activePanel, setActivePanel] = useState<Panel>('servers');

    useEffect(() => {
        Init().then((result: main.InitResult) => {
            switch (result.status) {
                case 'ready':
                    setLoading(false);
                    break;
                case 'fresh':
                    InitFresh().then(() => setLoading(false)).catch(console.error);
                    break;
                case 'needs_recovery':
                    setShowRecovery(true);
                    setLoading(false);
                    break;
            }
        }).catch(console.error);
    }, []);

    async function handleRecover() {
        await Recover();
        setShowRecovery(false);
    }

    async function handleReset() {
        await ResetAndInit();
        setShowRecovery(false);
    }

    if (loading) {
        return <div className="loading">Initializing…</div>;
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
                <SmartRecoveryModal onRecover={handleRecover} onReset={handleReset}/>
            )}
        </div>
    );
}

export default App;
