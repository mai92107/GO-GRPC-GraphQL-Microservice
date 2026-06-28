type Props = {
  activeBank: string;
  bankOptions: string[];
  cardCount: (bankName: string) => number;
  onChange: (bankName: string) => void;
};

export function ActivityBankTabs({
  activeBank,
  bankOptions,
  cardCount,
  onChange,
}: Props) {
  return (
    <div className="bank-tabs" role="tablist" aria-label="銀行">
      {bankOptions.map((bankName) => (
        <button
          type="button"
          role="tab"
          aria-selected={activeBank === bankName}
          className={`bank-tab ${activeBank === bankName ? "selected" : ""}`}
          key={bankName}
          onClick={() => onChange(bankName)}
        >
          <span>{bankName.slice(0, 2)}</span>
          <strong>{bankName}</strong>
          <small>{cardCount(bankName)} 張卡片</small>
        </button>
      ))}
    </div>
  );
}
