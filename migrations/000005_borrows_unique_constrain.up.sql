CREATE UNIQUE INDEX borrow_records_one_active ON borrow_records (user_id, book_id)
WHERE
  returned_at IS NULL;
