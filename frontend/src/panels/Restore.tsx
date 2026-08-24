import {useEffect, useState} from 'react';
import {main} from '../../wailsjs/go/models';
import {Analyze, GetCatalogObjects} from '../../wailsjs/go/main/App';
import {EventsOn} from '../../wailsjs/runtime/runtime';
import '../restore.css';

interface ScanProgress {
    percentBytes: number;
    databasesFound: number;
    tablesFound: number;
    blocksFound: number;
}

export function RestorePanel() {
    const [dumpFile, setDumpFile] = useState('');
    const [scanning, setScanning] = useState(false);
    const [progress, setProgress] = useState<ScanProgress | null>(null);
    const [objects, setObjects] = useState<main.CatalogObject[]>([]);
    const [selected, setSelected] = useState<Set<number>>(new Set());
    const [error, setError] = useState('');

    useEffect(() => {
        const unsub = EventsOn('scan:progress', (p: ScanProgress) => {
            setProgress(p);
        });
        return () => unsub();
    }, []);

    async function handleAnalyze() {
        if (!dumpFile) return;
        setScanning(true);
        setProgress(null);
        setObjects([]);
        setSelected(new Set());
        setError('');
        try {
            await Analyze(dumpFile);
            const objs = await GetCatalogObjects(dumpFile);
            setObjects(objs);
        } catch (e: unknown) {
            setError(String(e));
        } finally {
            setScanning(false);
        }
    }

    function toggleSelect(id: number) {
        setSelected((prev) => {
            const next = new Set(prev);
            if (next.has(id)) next.delete(id);
            else next.add(id);
            return next;
        });
    }

    function toggleDatabase(dbName: string) {
        const dbObjs = objects.filter((o) => o.databaseName === dbName && o.objectType === 'table');
        const allSelected = dbObjs.every((o) => selected.has(o.id));
        setSelected((prev) => {
            const next = new Set(prev);
            dbObjs.forEach((o) => {
                if (allSelected) next.delete(o.id);
                else next.add(o.id);
            });
            return next;
        });
    }

    // Group objects by database
    const databases = [...new Set(objects.map((o) => o.databaseName))];

    return (
        <div className="panel restore-panel">
            <h1>♻️ Restore</h1>

            <div className="restore-file-row">
                <input
                    type="text"
                    placeholder="Path to dump file (.sql)"
                    value={dumpFile}
                    onChange={(e) => setDumpFile(e.target.value)}
                />
                <button
                    className="btn btn-primary"
                    onClick={handleAnalyze}
                    disabled={!dumpFile || scanning}
                >
                    {scanning ? '⏳ Scanning…' : '🔍 Analyze'}
                </button>
            </div>

            {error && <div className="test-result error">{error}</div>}

            {scanning && progress && (
                <div className="scan-progress">
                    <div className="progress-bar">
                        <div className="progress-fill" style={{width: `${progress.percentBytes}%`}}/>
                    </div>
                    <div className="progress-stats">
                        {progress.percentBytes.toFixed(1)}% — {progress.databasesFound} DBs,{' '}
                        {progress.tablesFound} tables, {progress.blocksFound} blocks
                    </div>
                </div>
            )}

            {objects.length > 0 && (
                <div className="catalog-checklist">
                    <div className="checklist-header">
                        <span>{objects.length} objects found</span>
                        <span>{selected.size} selected</span>
                    </div>
                    {databases.map((db) => {
                        const dbTables = objects.filter(
                            (o) => o.databaseName === db && o.objectType === 'table'
                        );
                        const dbBlocks = objects.filter(
                            (o) => o.databaseName === db && o.objectType !== 'table'
                        );
                        const allTablesSelected = dbTables.every((o) => selected.has(o.id));
                        return (
                            <div key={db} className="db-group">
                                <div className="db-header">
                                    <input
                                        type="checkbox"
                                        checked={allTablesSelected && dbTables.length > 0}
                                        onChange={() => toggleDatabase(db)}
                                    />
                                    <span className="db-name">🗄️ {db}</span>
                                    <span className="db-count">{dbTables.length} tables</span>
                                </div>
                                <div className="db-objects">
                                    {dbTables.map((o) => (
                                        <label key={o.id} className="checklist-item">
                                            <input
                                                type="checkbox"
                                                checked={selected.has(o.id)}
                                                onChange={() => toggleSelect(o.id)}
                                            />
                                            <span className="obj-name">{o.objectName}</span>
                                            <span className="obj-bytes">
                                                {formatBytes(o.endByte - o.startByte)}
                                            </span>
                                        </label>
                                    ))}
                                    {dbBlocks.map((o) => (
                                        <div key={o.id} className="checklist-item block-item">
                                            <span className="obj-type">{o.objectType.replace('_block', '')}</span>
                                            <span className="obj-bytes">
                                                {formatBytes(o.endByte - o.startByte)}
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}

            {objects.length === 0 && !scanning && !error && (
                <div className="panel-placeholder">
                    Select a dump file and click Analyze to scan its contents.
                </div>
            )}
        </div>
    );
}

function formatBytes(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
