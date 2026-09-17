export default function OptionBar({ option, count, total, onVote, disabled, isWinner }) {
  const pct = total > 0 ? Math.round((count / total) * 100) : 0

  return (
    <div className={`option-row ${isWinner ? 'option-row-leading' : ''}`}>
      <button
        className="option-row-hit"
        onClick={() => onVote?.(option.id)}
        disabled={disabled || !onVote}
      >
        <div className="option-row-track">
          <div className="option-row-fill" style={{ width: `${pct}%` }} />
        </div>
        <div className="option-row-meta">
          <span className="option-row-text">{option.text}</span>
          <span className="option-row-stats">
            <span className="option-row-pct">{pct}%</span>
            <span className="option-row-count">{count}</span>
          </span>
        </div>
      </button>
    </div>
  )
}
