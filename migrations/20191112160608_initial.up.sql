CREATE TABLE IF NOT EXISTS todos
(
  id serial PRIMARY KEY,
  title character varying (2047) NOT NULL,
  due_date timestamp without time zone NOT NULL,
  done boolean NOT NULL
);

CREATE TABLE IF NOT EXISTS comments
(
  id serial PRIMARY KEY,
  text character varying (2047) NOT NULL,
  todo_id integer REFERENCES todos(id) NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx__todo_comments__todo_id ON comments(todo_id);

CREATE TABLE IF NOT EXISTS labels
(
  id serial PRIMARY KEY,
  text character varying (2047) NOT NULL,
  color character varying (255) NOT NULL,
  todo_id integer REFERENCES todos(id) NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx__todo_labels__todo_id ON labels(todo_id);