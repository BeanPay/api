CREATE TYPE billfrequency AS ENUM ('monthly', 'quarterly', 'biannually', 'annually');

CREATE TABLE bills(
  id                    uuid            PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id               uuid            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name                  text            NOT NULL,
  payment_url           text            NOT NULL,
  frequency             billfrequency   NOT NULL,
  estimated_total_due   NUMERIC(8, 2)   NOT NULL,
  first_due_date        date            NOT NULL,
  created_at            timestamptz     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at            timestamptz     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER bills_updated_at
BEFORE UPDATE ON bills
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
