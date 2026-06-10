export default function SidebarSection({ eyebrow, title, actionLabel, onAction, children, className = "", sectionRef }) {
  return (
    <section ref={sectionRef} className={`panel-section ${className}`.trim()}>
      <div className="panel-section-header">
        <div>
          {eyebrow ? <div className="panel-eyebrow">{eyebrow}</div> : null}
          <h2>{title}</h2>
        </div>
        {actionLabel ? (
          <button className="ghost-button" type="button" onClick={onAction}>
            {actionLabel}
          </button>
        ) : null}
      </div>
      {children}
    </section>
  );
}
