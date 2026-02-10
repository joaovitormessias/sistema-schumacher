interface ProgressBarProps {
    value: number;
    max?: number;
    className?: string;
    showLabel?: boolean;
}

export function ProgressBar({ value, max = 100, className = "", showLabel = false }: ProgressBarProps) {
    const percentage = Math.min(Math.max((value / max) * 100, 0), 100);

    return (
        <div className={`progress-bar-container ${className}`}>
            <div className="progress-bar">
                <div className="progress-bar-fill" style={{ width: `${percentage}%` }} />
            </div>
            {showLabel && <span className="progress-bar-label">{Math.round(percentage)}%</span>}
        </div>
    );
}

export default ProgressBar;
