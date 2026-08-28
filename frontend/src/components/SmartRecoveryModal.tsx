interface SmartRecoveryModalProps {
    onCancel: () => void;
    onReset: () => void;
}

export function SmartRecoveryModal({onCancel, onReset}: SmartRecoveryModalProps) {
    return (
        <div className="modal-overlay">
            <div className="modal">
                <h2>⚠️ Smart Recovery</h2>
                <p>
                    The encryption key (<code>app.key</code>) is missing, but an existing
                    database catalog was found.
                </p>
                <p>How would you like to proceed?</p>
                <div className="modal-actions">
                    <button className="btn btn-secondary" onClick={onCancel}>
                        Cancel — I'll recover the key
                    </button>
                    <button className="btn btn-danger" onClick={onReset}>
                        Reset &amp; Re-init
                    </button>
                </div>
            </div>
        </div>
    );
}
