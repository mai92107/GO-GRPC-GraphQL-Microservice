
export default function Nav(
  {icon, label, active, onClick}: 
  {
    icon: React.ReactNode;
    label: string;
    active: boolean;
    onClick: () => void;
  }) {
      return (
        <button className={active ? "active" : ""} onClick={onClick}>
          {icon}
          {label}
        </button>
      );
    }