CREATE TABLE IF NOT EXISTS books (
  id bigserial PRIMARY KEY,
  title text NOT NULL,
  author text NOT NULL,
  publish_date DATE NOT NULL,
  isbn VARCHAR(17) UNIQUE NOT NULL,
  description text NOT NULL,
  cover_image text NOT NULL,
  genres text[] NOT NULL,
  pages integer NOT NULL DEFAULT 0,
  language VARCHAR(50) NOT NULL DEFAULT 'English',
  publisher text NOT NULL,
  copies_total integer NOT NULL DEFAULT 1,
  copies_available integer NOT NULL DEFAULT 1,
  version integer NOT NULL DEFAULT 1
);
