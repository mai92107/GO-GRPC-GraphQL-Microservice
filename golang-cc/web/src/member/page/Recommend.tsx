import { ConfirmationDialog } from "./recommend/ConfirmationDialog";
import { RecommendForm } from "./recommend/RecommendForm";
import { RecommendationResults } from "./recommend/RecommendationResults";
import { useRecommend } from "./recommend/useRecommend";

export function Recommend() {
  const recommend = useRecommend();

  return (
    <>
      <div className="page-head">
        <div>
          <span className="page-eyebrow">SMART RECOMMENDATION</span>
          <h1>這次刷哪張？</h1>
          <p>輸入金額與消費情境，立即找到回饋最高的卡片與支付方式。</p>
        </div>
      </div>
      {recommend.notice && <div className="notice">{recommend.notice}</div>}
      <div className="grid">
        <RecommendForm
          amount={recommend.amount}
          categories={recommend.availableCategories}
          category={recommend.category}
          date={recommend.date}
          error={recommend.error}
          loadingCategories={recommend.loadingCategories}
          loadingMerchants={recommend.loadingMerchants}
          merchantId={recommend.merchantId}
          merchants={recommend.merchants}
          onAmount={recommend.setAmount}
          onCategory={recommend.setCategory}
          onDate={recommend.setDate}
          onMerchant={recommend.setMerchantId}
          onOtherMerchant={recommend.setOtherMerchant}
          onSubmit={recommend.run}
          otherMerchant={recommend.otherMerchant}
          recommending={recommend.recommending}
        />
        <RecommendationResults
          hasSearched={recommend.hasSearched}
          onConfirm={recommend.setConfirmation}
          results={recommend.results}
        />
      </div>
      {recommend.confirmation && (
        <ConfirmationDialog
          amount={recommend.amount}
          confirmation={recommend.confirmation}
          merchantName={recommend.merchantName}
          onChange={recommend.setConfirmation}
          onClose={() => !recommend.recording && recommend.setConfirmation(null)}
          onTransact={() =>
            recommend.confirmation &&
            void recommend.transact(
              recommend.confirmation.card,
              recommend.confirmation.option,
              recommend.confirmation.paymentCode,
            )
          }
          recording={recommend.recording}
        />
      )}
    </>
  );
}
