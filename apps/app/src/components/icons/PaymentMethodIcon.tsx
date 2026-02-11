import React from "react";

type PaymentMethod = "PIX" | "CARD" | "CASH" | "TRANSFER" | "OTHER";
type Size = 16 | 24 | 32 | 48;

interface PaymentMethodIconProps {
  method: PaymentMethod;
  size?: Size;
  color?: string;
}

const sizeMap: Record<Size, number> = {
  16: 16,
  24: 24,
  32: 32,
  48: 48,
};

export default function PaymentMethodIcon({
  method,
  size = 24,
  color = "currentColor",
}: PaymentMethodIconProps) {
  const dimension = sizeMap[size];
  const scale = size / 24;

  const PIXIcon = (
    <svg
      width={dimension}
      height={dimension}
      viewBox="0 0 24 24"
      fill="none"
      stroke={color}
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      {/* QR Code Grid */}
      <rect x="2" y="2" width="20" height="20" rx="2" />
      <rect x="4" y="4" width="6" height="6" />
      <rect x="14" y="4" width="6" height="6" />
      <rect x="4" y="14" width="6" height="6" />
      <circle cx="17" cy="17" r="2.5" fill={color} />
      {/* Timing marks */}
      <line x1="11" y1="4" x2="11" y2="10" stroke={color} strokeWidth="1" />
      <line x1="4" y1="11" x2="10" y2="11" stroke={color} strokeWidth="1" />
    </svg>
  );

  const CARDIcon = (
    <svg
      width={dimension}
      height={dimension}
      viewBox="0 0 24 24"
      fill="none"
      stroke={color}
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      {/* Card background */}
      <rect x="1" y="4" width="22" height="16" rx="2" ry="2" />
      {/* Magnetic stripe */}
      <line x1="1" y1="10" x2="23" y2="10" stroke={color} strokeWidth="3" />
      {/* Chip */}
      <rect x="2" y="11" width="4" height="4" rx="1" />
    </svg>
  );

  const CASHIcon = (
    <svg
      width={dimension}
      height={dimension}
      viewBox="0 0 24 24"
      fill="none"
      stroke={color}
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      {/* Dollar sign */}
      <path d="M12 1v22M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
    </svg>
  );

  const TRANSFERIcon = (
    <svg
      width={dimension}
      height={dimension}
      viewBox="0 0 24 24"
      fill="none"
      stroke={color}
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      {/* Bank building */}
      <path d="M3 11h18M3 11v8a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-8" />
      <path d="M12 3L2 9h20Z" />
      {/* Vertical lines for columns */}
      <line x1="8" y1="11" x2="8" y2="19" stroke={color} />
      <line x1="12" y1="11" x2="12" y2="19" stroke={color} />
      <line x1="16" y1="11" x2="16" y2="19" stroke={color} />
    </svg>
  );

  const OTHERIcon = (
    <svg
      width={dimension}
      height={dimension}
      viewBox="0 0 24 24"
      fill="none"
      stroke={color}
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      {/* Wallet icon */}
      <path d="M19 7V4a1 1 0 0 0-1-1H5a2 2 0 0 0 0 4h15a1 1 0 0 1 1 1v4H4a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-4a1 1 0 0 0-1-1z" />
      <circle cx="16" cy="13" r="1.5" fill={color} />
    </svg>
  );

  const icons: Record<PaymentMethod, React.ReactNode> = {
    PIX: PIXIcon,
    CARD: CARDIcon,
    CASH: CASHIcon,
    TRANSFER: TRANSFERIcon,
    OTHER: OTHERIcon,
  };

  return <>{icons[method]}</>;
}
