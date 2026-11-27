/** Tone colors by number, matching Pinyin's tone coloring. */
const stroke = ["", "#C8102E", "#B45309", "#15803D", "#1D4ED8", "#6B7280"];

/**
 * A tone's pitch path on the five-level scale (5 high, 1 low), drawn left to right. With no
 * contour (the neutral tone) it shows a short dot at mid height.
 */
export function ToneContour({ contour, tone, className = "h-16 w-24", grid = true }: { contour?: number[]; tone: number; className?: string; grid?: boolean }) {
  const W = 120;
  const H = 80;
  const y = (level: number) => H - 8 - ((level - 1) / 4) * (H - 16);
  const color = stroke[tone] ?? "#1C1917";
  const points = contour && contour.length > 1 ? contour.map((level, i) => [12 + (i / (contour.length - 1)) * (W - 24), y(level)]) : null;
  const path = points
    ? points.length === 3
      ? `M${points[0][0]},${points[0][1]} Q${points[1][0]},${points[1][1] + 14} ${points[2][0]},${points[2][1]}`
      : `M${points[0][0]},${points[0][1]} L${points[1][0]},${points[1][1]}`
    : "";
  const end = points?.[points.length - 1];
  return (
    <svg viewBox={`0 0 ${W} ${H}`} className={className} role="img" aria-label={contour ? `Pitch ${contour.join("")}` : "Short, light, no tone of its own"}>
      {grid &&
        [1, 2, 3, 4, 5].map((level) => (
          <line key={level} x1="4" x2={W - 4} y1={y(level)} y2={y(level)} stroke="#E7E2DC" strokeWidth={level === 3 ? 1.5 : 1} strokeDasharray={level === 3 ? "" : "3 4"} />
        ))}
      {points ? (
        <>
          <path d={path} fill="none" stroke={color} strokeWidth="7" strokeLinecap="round" />
          {end && <circle cx={end[0]} cy={end[1]} r="7" fill={color} />}
        </>
      ) : (
        <circle cx={W / 2} cy={y(3)} r="8" fill={color} />
      )}
    </svg>
  );
}
