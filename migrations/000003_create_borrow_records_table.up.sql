CREATE TABLE IF NOT EXISTS borrow_records (
  id bigserial PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  book_id bigint NOT NULL REFERENCES books (id) ON DELETE CASCADE,
  borrowed_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  returned_at timestamp(0) with time zone NULL
);
