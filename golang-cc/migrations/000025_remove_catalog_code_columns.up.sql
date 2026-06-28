CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'merchants'
          AND column_name = 'code'
    ) THEN
        ALTER TABLE merchants ADD COLUMN IF NOT EXISTS id UUID;
        UPDATE merchants SET id = gen_random_uuid() WHERE id IS NULL;
        ALTER TABLE merchants ALTER COLUMN id SET DEFAULT gen_random_uuid();
        ALTER TABLE merchants ALTER COLUMN id SET NOT NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'category_code')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'category_id') THEN
        ALTER TABLE transactions RENAME COLUMN category_code TO category_id;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'payment_method_code')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'payment_method_id') THEN
        ALTER TABLE transactions RENAME COLUMN payment_method_code TO payment_method_id;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'currency_code')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'currency_id') THEN
        ALTER TABLE transactions RENAME COLUMN currency_code TO currency_id;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'country_code')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'country_id') THEN
        ALTER TABLE transactions RENAME COLUMN country_code TO country_id;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'member_profile' AND table_name = 'user_payment_methods' AND column_name = 'payment_method_code')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'member_profile' AND table_name = 'user_payment_methods' AND column_name = 'payment_method_id') THEN
        ALTER TABLE member_profile.user_payment_methods RENAME COLUMN payment_method_code TO payment_method_id;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'merchant_categories' AND column_name = 'category_code')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'merchant_categories' AND column_name = 'category_id') THEN
        ALTER TABLE merchant_categories RENAME COLUMN category_code TO category_id;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'categories' AND column_name = 'code')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'categories' AND column_name = 'id') THEN
        ALTER TABLE categories RENAME COLUMN code TO id;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'payment_methods' AND column_name = 'code')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'payment_methods' AND column_name = 'id') THEN
        ALTER TABLE payment_methods RENAME COLUMN code TO id;
    END IF;
END $$;

UPDATE reward.condition_versions
SET configuration_json = configuration_json - 'category_codes' ||
    jsonb_build_object('category_ids', configuration_json->'category_codes')
WHERE configuration_json ? 'category_codes';

UPDATE reward.condition_versions
SET configuration_json = configuration_json - 'payment_method_codes' ||
    jsonb_build_object('payment_method_ids', configuration_json->'payment_method_codes')
WHERE configuration_json ? 'payment_method_codes';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'merchants'
          AND column_name = 'code'
    ) THEN
        EXECUTE $update$
            UPDATE reward.condition_versions
            SET configuration_json = configuration_json - 'merchant_codes' ||
                jsonb_build_object(
                    'merchant_ids',
                    COALESCE((
                        SELECT jsonb_agg(m.id::text ORDER BY merchant_values.ordinality)
                        FROM jsonb_array_elements_text(configuration_json->'merchant_codes') WITH ORDINALITY AS merchant_values(value, ordinality)
                        JOIN merchants m ON m.code = merchant_values.value
                    ), '[]'::jsonb)
                )
            WHERE configuration_json ? 'merchant_codes'
        $update$;
    ELSE
        UPDATE reward.condition_versions
        SET configuration_json = configuration_json - 'merchant_codes' ||
            jsonb_build_object('merchant_ids', configuration_json->'merchant_codes')
        WHERE configuration_json ? 'merchant_codes';
    END IF;
END $$;

DO $$
DECLARE
    constraint_record RECORD;
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'merchants'
          AND column_name = 'code'
    ) THEN
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'merchant_aliases' AND column_name = 'merchant_code'
        ) THEN
            ALTER TABLE merchant_aliases ADD COLUMN IF NOT EXISTS merchant_id UUID;
            UPDATE merchant_aliases a SET merchant_id = m.id FROM merchants m WHERE a.merchant_code = m.code;
            ALTER TABLE merchant_aliases ALTER COLUMN merchant_id SET NOT NULL;
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'merchant_categories' AND column_name = 'merchant_code'
        ) THEN
            ALTER TABLE merchant_categories ADD COLUMN IF NOT EXISTS merchant_id UUID;
            UPDATE merchant_categories c SET merchant_id = m.id FROM merchants m WHERE c.merchant_code = m.code;
            ALTER TABLE merchant_categories ALTER COLUMN merchant_id SET NOT NULL;
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'card_activity_benefit_merchants' AND column_name = 'merchant_code'
        ) THEN
            ALTER TABLE card_activity_benefit_merchants ADD COLUMN IF NOT EXISTS merchant_id UUID;
            UPDATE card_activity_benefit_merchants bm SET merchant_id = m.id FROM merchants m WHERE bm.merchant_code = m.code;
            ALTER TABLE card_activity_benefit_merchants ALTER COLUMN merchant_id SET NOT NULL;
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'merchant_code'
        ) THEN
            ALTER TABLE transactions ADD COLUMN IF NOT EXISTS merchant_id UUID;
            UPDATE transactions t SET merchant_id = m.id FROM merchants m WHERE t.merchant_code = m.code;
        END IF;

        FOR constraint_record IN
            SELECT conrelid::regclass AS table_name, conname
            FROM pg_constraint
            WHERE conrelid IN (
                SELECT table_oid
                FROM unnest(ARRAY[
                    to_regclass('merchants'),
                    to_regclass('merchant_aliases'),
                    to_regclass('merchant_categories'),
                    to_regclass('card_activity_benefit_merchants'),
                    to_regclass('transactions')
                ]) AS existing_tables(table_oid)
                WHERE table_oid IS NOT NULL
            )
            AND (
                confrelid = 'merchants'::regclass
                OR pg_get_constraintdef(oid) LIKE '%code%'
            )
            ORDER BY
                CASE WHEN conrelid = 'merchants'::regclass AND contype = 'p' THEN 1 ELSE 0 END,
                conrelid::regclass::text,
                conname
        LOOP
            EXECUTE format('ALTER TABLE %s DROP CONSTRAINT IF EXISTS %I', constraint_record.table_name, constraint_record.conname);
        END LOOP;

        ALTER TABLE merchants DROP COLUMN code;
        ALTER TABLE merchants ADD PRIMARY KEY (id);

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'merchant_aliases' AND column_name = 'merchant_code'
        ) THEN
            ALTER TABLE merchant_aliases DROP COLUMN merchant_code;
            ALTER TABLE merchant_aliases ADD PRIMARY KEY (merchant_id, alias);
            ALTER TABLE merchant_aliases ADD FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE;
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'merchant_categories' AND column_name = 'merchant_code'
        ) THEN
            ALTER TABLE merchant_categories DROP COLUMN merchant_code;
            ALTER TABLE merchant_categories ADD PRIMARY KEY (merchant_id, category_id);
            ALTER TABLE merchant_categories ADD FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE;
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'card_activity_benefit_merchants' AND column_name = 'merchant_code'
        ) THEN
            ALTER TABLE card_activity_benefit_merchants DROP COLUMN merchant_code;
            ALTER TABLE card_activity_benefit_merchants ADD PRIMARY KEY (benefit_id, merchant_id);
            ALTER TABLE card_activity_benefit_merchants ADD FOREIGN KEY (merchant_id) REFERENCES merchants(id);
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'merchant_code'
        ) THEN
            DROP INDEX IF EXISTS transactions_merchant_code_idx;
            ALTER TABLE transactions DROP COLUMN merchant_code;
            CREATE INDEX IF NOT EXISTS transactions_merchant_id_idx ON transactions(merchant_id);
            ALTER TABLE transactions ADD FOREIGN KEY (merchant_id) REFERENCES merchants(id);
        END IF;
    END IF;
END $$;

ALTER TABLE reward_units DROP COLUMN IF EXISTS code;
ALTER TABLE card_products DROP COLUMN IF EXISTS code CASCADE;
ALTER TABLE catalog.card_networks DROP COLUMN IF EXISTS code CASCADE;
ALTER TABLE catalog.card_plans DROP COLUMN IF EXISTS code CASCADE;
ALTER TABLE reward.programs DROP COLUMN IF EXISTS code CASCADE;
ALTER TABLE reward.components DROP COLUMN IF EXISTS code CASCADE;
ALTER TABLE reward.caps DROP COLUMN IF EXISTS code CASCADE;
ALTER TABLE reward.conditions DROP COLUMN IF EXISTS code CASCADE;
