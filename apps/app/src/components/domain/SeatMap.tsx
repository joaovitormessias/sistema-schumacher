import { useMemo } from "react";
import { Star } from "lucide-react";

export interface Seat {
    id: string;
    number: number;
    status: "available" | "occupied" | "selected";
    isPreferred?: boolean;
}

interface SeatMapProps {
    seats: Seat[];
    selectedSeatId?: string;
    preferredSeatIds?: string[];
    seatsPerRow?: number;
    onSeatClick?: (seat: Seat) => void;
    showLegend?: boolean;
    className?: string;
}

export function SeatMap({
    seats,
    selectedSeatId,
    preferredSeatIds = [],
    seatsPerRow = 4,
    onSeatClick,
    showLegend = true,
    className = "",
}: SeatMapProps) {
    // Group seats into rows (2 seats | aisle | 2 seats pattern)
    const rows = useMemo(() => {
        const result: Seat[][] = [];
        for (let i = 0; i < seats.length; i += seatsPerRow) {
            result.push(seats.slice(i, i + seatsPerRow));
        }
        return result;
    }, [seats, seatsPerRow]);

    const getSeatClasses = (seat: Seat) => {
        const classes = ["seat"];

        if (seat.id === selectedSeatId) {
            classes.push("seat-selected");
        } else if (seat.status === "occupied") {
            classes.push("seat-occupied");
        } else {
            classes.push("seat-available");
        }

        if (preferredSeatIds.includes(seat.id) && seat.status !== "occupied") {
            classes.push("seat-preferred");
        }

        return classes.join(" ");
    };

    const handleSeatClick = (seat: Seat) => {
        if (seat.status === "occupied") return;
        onSeatClick?.(seat);
    };

    return (
        <div className={`seat-map ${className}`}>
            {rows.map((row, rowIdx) => (
                <div key={rowIdx} className="seat-map-row">
                    {row.map((seat, seatIdx) => (
                        <>
                            {/* Add aisle after first half of seats */}
                            {seatIdx === Math.floor(seatsPerRow / 2) && <div className="seat-map-aisle" />}
                            <button
                                key={seat.id}
                                className={getSeatClasses(seat)}
                                onClick={() => handleSeatClick(seat)}
                                disabled={seat.status === "occupied"}
                                aria-label={`Poltrona ${seat.number}${seat.status === "occupied" ? " (ocupada)" : ""}${preferredSeatIds.includes(seat.id) ? " (preferida)" : ""
                                    }`}
                                title={`Poltrona ${seat.number}`}
                            >
                                {seat.number}
                                {preferredSeatIds.includes(seat.id) && seat.status !== "occupied" && (
                                    <Star
                                        size={10}
                                        fill="#fbbf24"
                                        color="#fbbf24"
                                        style={{ position: "absolute", top: 2, right: 2 }}
                                    />
                                )}
                            </button>
                        </>
                    ))}
                </div>
            ))}

            {showLegend && (
                <div className="seat-legend">
                    <div className="seat-legend-item">
                        <div className="seat-legend-dot available" />
                        <span>Disponível</span>
                    </div>
                    <div className="seat-legend-item">
                        <div className="seat-legend-dot occupied" />
                        <span>Ocupada</span>
                    </div>
                    <div className="seat-legend-item">
                        <div className="seat-legend-dot selected" />
                        <span>Selecionada</span>
                    </div>
                    <div className="seat-legend-item">
                        <div className="seat-legend-dot preferred" />
                        <span>Preferida</span>
                    </div>
                </div>
            )}
        </div>
    );
}

export default SeatMap;
