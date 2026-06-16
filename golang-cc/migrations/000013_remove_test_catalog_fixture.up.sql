DELETE FROM card_activities
WHERE id='30000000-0000-0000-0000-000000000002'
  AND name='測試活動';

DELETE FROM card_products
WHERE bank_id='40000000-0000-0000-0000-000000000001'
  AND NOT EXISTS (
      SELECT 1 FROM member_cards mc WHERE mc.card_product_id=card_products.id
  )
  AND NOT EXISTS (
      SELECT 1 FROM card_activities a WHERE a.card_product_id=card_products.id
  );

DELETE FROM banks
WHERE id='40000000-0000-0000-0000-000000000001'
  AND name='虛構銀行'
  AND NOT EXISTS (
      SELECT 1 FROM card_products cp WHERE cp.bank_id=banks.id
  );
