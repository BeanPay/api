CREATE TABLE payments(
  id            uuid            PRIMARY KEY DEFAULT gen_random_uuid(),
  bill_id       uuid            NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
  due_date      date            NOT NULL,
  total_paid    NUMERIC(8, 2)   NOT NULL,
  created_at    timestamptz     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    timestamptz     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(bill_id, due_date)
);

CREATE TRIGGER payments_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

/* Index bill_id as almost all of our read queries against this
 * table will be against a list of bill_ids. */
CREATE INDEX bill_id_idx ON payments (bill_id);
