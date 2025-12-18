ALTER TABLE borrow_records
ADD COLUMN due_at timestamp(0) with time zone NOT NULL;
