CREATE TABLE "nurse" (
  "id" INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "username" varchar UNIQUE,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "hashed_password" varchar NOT NULL,
 "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01--01 00:00:00Z',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

