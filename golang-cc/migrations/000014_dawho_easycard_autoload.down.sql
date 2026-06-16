DELETE FROM card_activity_benefits
WHERE id IN (
    '71000000-0000-0000-0000-000000000022',
    '71000000-0000-0000-0000-000000000023'
);

UPDATE card_activities
SET source_url = 'https://bank.sinopac.com/sinopacBT/webevents/dawho/about/card/',
    verified_at = DATE '2026-06-15'
WHERE id = '61000000-0000-0000-0000-000000000001';
