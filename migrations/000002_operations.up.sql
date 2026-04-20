CREATE TABLE operations (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL,
  amount NUMERIC,
  category TEXT,
  created_at TIMESTAMP DEFAULT NOW()
)
