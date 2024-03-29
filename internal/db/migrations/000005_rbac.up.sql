CREATE TABLE "roles" (
  "roleid" INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "role" varchar UNIQUE NOT NULL
);

CREATE TABLE "permissions" (
  "permissionid" INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "permission" varchar NOT NULL,
  "roleid" integer NOT NULL
);

CREATE TABLE "users" (
  "id" INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "password" varchar NOT NULL,
  "roleid" integer NOT NULL
);

CREATE INDEX ON "roles" ("roleid");

CREATE INDEX ON "permissions" ("permissionid");

CREATE INDEX ON "users" ("id");

ALTER TABLE "permissions" ADD FOREIGN KEY ("roleid") REFERENCES "roles" ("roleid") ON DELETE CASCADE ;

ALTER TABLE "users" ADD FOREIGN KEY ("roleid") REFERENCES "roles" ("roleid") ON DELETE CASCADE;

