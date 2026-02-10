import type { CSSProperties, ReactNode } from "react";

interface SkeletonTextProps {
    lines?: number;
    className?: string;
}

export function SkeletonText({ lines = 2, className = "" }: SkeletonTextProps) {
    return (
        <div className={className}>
            {Array.from({ length: lines }).map((_, i) => (
                <div key={i} className="skeleton skeleton-text" />
            ))}
        </div>
    );
}

interface SkeletonCircleProps {
    size?: number;
    className?: string;
}

export function SkeletonCircle({ size = 40, className = "" }: SkeletonCircleProps) {
    const style: CSSProperties = { width: size, height: size };
    return <div className={`skeleton skeleton-circle ${className}`} style={style} />;
}

interface SkeletonButtonProps {
    width?: number | string;
    className?: string;
}

export function SkeletonButton({ width = 120, className = "" }: SkeletonButtonProps) {
    const style: CSSProperties = { width };
    return <div className={`skeleton skeleton-button ${className}`} style={style} />;
}

interface SkeletonCardProps {
    children?: ReactNode;
    className?: string;
}

export function SkeletonCard({ children, className = "" }: SkeletonCardProps) {
    return (
        <div className={`skeleton-card ${className}`}>
            {children ?? (
                <>
                    <SkeletonCircle size={48} />
                    <SkeletonText lines={2} />
                    <SkeletonButton />
                </>
            )}
        </div>
    );
}

interface SkeletonTableProps {
    rows?: number;
    columns?: number;
    className?: string;
}

export function SkeletonTable({ rows = 5, columns = 4, className = "" }: SkeletonTableProps) {
    const style: CSSProperties = {
        "--table-columns": `repeat(${columns}, 1fr)`,
    } as CSSProperties;

    return (
        <div className={className}>
            {Array.from({ length: rows }).map((_, rowIdx) => (
                <div key={rowIdx} className="skeleton-table-row" style={style}>
                    {Array.from({ length: columns }).map((_, colIdx) => (
                        <div key={colIdx} className="skeleton-table-cell" />
                    ))}
                </div>
            ))}
        </div>
    );
}

interface SkeletonStatCardProps {
    className?: string;
}

export function SkeletonStatCard({ className = "" }: SkeletonStatCardProps) {
    return (
        <div className={`stat-card ${className}`}>
            <div className="skeleton skeleton-text" style={{ width: "60%", height: 12 }} />
            <div className="skeleton skeleton-text" style={{ width: "40%", height: 28, marginTop: 8 }} />
            <div className="skeleton skeleton-text" style={{ width: "80%", height: 12, marginTop: 8 }} />
        </div>
    );
}

// Compound component for flexible skeleton composition
export const Skeleton = {
    Text: SkeletonText,
    Circle: SkeletonCircle,
    Button: SkeletonButton,
    Card: SkeletonCard,
    Table: SkeletonTable,
    StatCard: SkeletonStatCard,
};

export default Skeleton;
