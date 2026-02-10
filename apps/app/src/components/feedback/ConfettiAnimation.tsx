import { useEffect, useState } from "react";

interface ConfettiPiece {
    id: number;
    left: number;
    delay: number;
    color: string;
    size: number;
}

interface ConfettiAnimationProps {
    count?: number;
    colors?: string[];
    duration?: number;
    onComplete?: () => void;
}

export function ConfettiAnimation({
    count = 50,
    colors = ["#0f766e", "#15803d", "#2563eb", "#fbbf24", "#ec4899"],
    duration = 3000,
    onComplete,
}: ConfettiAnimationProps) {
    const [pieces, setPieces] = useState<ConfettiPiece[]>([]);
    const [visible, setVisible] = useState(true);

    useEffect(() => {
        // Generate confetti pieces
        const newPieces: ConfettiPiece[] = Array.from({ length: count }).map((_, i) => ({
            id: i,
            left: Math.random() * 100,
            delay: Math.random() * 0.5,
            color: colors[Math.floor(Math.random() * colors.length)],
            size: 6 + Math.random() * 8,
        }));

        setPieces(newPieces);

        // Clean up after animation completes
        const timer = setTimeout(() => {
            setVisible(false);
            onComplete?.();
        }, duration);

        return () => clearTimeout(timer);
    }, [count, colors, duration, onComplete]);

    if (!visible) return null;

    return (
        <div className="confetti-container" aria-hidden="true">
            {pieces.map((piece) => (
                <div
                    key={piece.id}
                    className="confetti-piece"
                    style={{
                        left: `${piece.left}%`,
                        animationDelay: `${piece.delay}s`,
                        backgroundColor: piece.color,
                        width: piece.size,
                        height: piece.size,
                        borderRadius: Math.random() > 0.5 ? "50%" : "2px",
                    }}
                />
            ))}
        </div>
    );
}

export default ConfettiAnimation;
