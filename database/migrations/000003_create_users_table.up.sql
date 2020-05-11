CREATE TABLE users(
  id            uuid          PRIMARY KEY DEFAULT gen_random_uuid(),
  email         text          NOT NULL UNIQUE,
  password      varchar(60)   NOT NULL,
  created_at    timestamptz   NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    timestamptz   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
