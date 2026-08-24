import {Panel} from '../App';

interface SidebarProps {
    active: Panel;
    onChange: (panel: Panel) => void;
}

const items: {id: Panel; label: string; icon: string}[] = [
    {id: 'servers', label: 'Servers', icon: '🖥️'},
    {id: 'backup', label: 'Backup', icon: '💾'},
    {id: 'restore', label: 'Restore', icon: '♻️'},
    {id: 'settings', label: 'Settings', icon: '⚙️'},
];

export function Sidebar({active, onChange}: SidebarProps) {
    return (
        <nav className="sidebar">
            <div className="sidebar-header">MariaDB</div>
            {items.map((item) => (
                <button
                    key={item.id}
                    className={`sidebar-item ${active === item.id ? 'active' : ''}`}
                    onClick={() => onChange(item.id)}
                >
                    <span className="sidebar-icon">{item.icon}</span>
                    {item.label}
                </button>
            ))}
        </nav>
    );
}
