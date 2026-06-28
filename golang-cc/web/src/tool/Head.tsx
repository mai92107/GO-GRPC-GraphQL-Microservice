import type { ReactNode } from "react";

export default function Head({
  title,
  text,
  eyebrow = "ADMIN CONSOLE",
  actions,
}: {
  title: string;
  text: string;
  eyebrow?: string;
  actions?: ReactNode;
}) {
  return (
    <div className="page-head">
      <div>
        <span className="page-eyebrow">{eyebrow}</span>
        <h1>{title}</h1>
        <p>{text}</p>
      </div>
      {actions && <div className="page-actions">{actions}</div>}
    </div>
  );
}
