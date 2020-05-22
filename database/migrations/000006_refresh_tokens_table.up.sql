CREATE TABLE refresh_tokens(
  id            uuid            PRIMARY KEY DEFAULT gen_random_uuid(),
  chain_id      uuid            NOT NULL,
  user_id       uuid            NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at    timestamptz     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

/* Index chain_id as we will run a lot of queries to check if
 * the refresh_token is the most recent in the chain to validate it */
CREATE INDEX chain_idx ON refresh_tokens(chain_id);
