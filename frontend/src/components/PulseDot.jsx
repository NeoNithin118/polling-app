export default function PulseDot({ label }) {
  return (
    <span className="pulse-wrap">
      <span className="pulse-dot" />
      {label && <span className="pulse-label">{label}</span>}
    </span>
  )
}
