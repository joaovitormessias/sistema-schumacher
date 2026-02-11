import React from "react";

type RuleType = "OCCUPANCY" | "LEAD_TIME" | "DOW" | "SEASON";
type Size = 16 | 20 | 24 | 32;

interface RuleTypeIconProps {
  ruleType: RuleType;
  size?: Size;
  color?: string;
}

const sizeMap: Record<Size, number> = {
  16: 16,
  20: 20,
  24: 24,
  32: 32,
};

export default function RuleTypeIcon({
  ruleType,
  size = 24,
  color = "currentColor",
}: RuleTypeIconProps) {
  const dimension = sizeMap[size];

  const OCCUPANCYIcon = (
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
      {/* People/crowd icon - three people */}
      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
      <path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
  );

  const LEADTIMEIcon = (
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
      {/* Clock icon */}
      <circle cx="12" cy="12" r="10" />
      <polyline points="12 6 12 12 16 14" />
    </svg>
  );

  const DOWIcon = (
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
      {/* Calendar/Week icon */}
      <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
      <path d="M16 2v4M8 2v4M3 10h18" />
      {/* Week days indicated by small marks */}
      <circle cx="6" cy="16" r="1.5" fill={color} />
      <circle cx="12" cy="16" r="1.5" fill={color} />
      <circle cx="18" cy="16" r="1.5" fill={color} />
      <circle cx="6" cy="20" r="1.5" fill={color} />
      <circle cx="12" cy="20" r="1.5" fill={color} />
    </svg>
  );

  const SEASONIcon = (
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
      {/* Sun icon for seasonal */}
      <circle cx="12" cy="12" r="5" />
      <line x1="12" y1="1" x2="12" y2="3" />
      <line x1="12" y1="21" x2="12" y2="23" />
      <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
      <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
      <line x1="1" y1="12" x2="3" y2="12" />
      <line x1="21" y1="12" x2="23" y2="12" />
      <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
      <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
    </svg>
  );

  const icons: Record<RuleType, React.ReactNode> = {
    OCCUPANCY: OCCUPANCYIcon,
    LEAD_TIME: LEADTIMEIcon,
    DOW: DOWIcon,
    SEASON: SEASONIcon,
  };

  return <>{icons[ruleType]}</>;
}
