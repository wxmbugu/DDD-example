CREATE TABLE IF NOT EXISTS "department" (
  "departmentid" SERIAL PRIMARY KEY,
  "departmentname" varchar UNIQUE NOT NULL
);

CREATE INDEX ON "department" ("departmentid");
ALTER TABLE IF EXISTS "physician" ADD COLUMN "departmentname" varchar  NOT NULL;
ALTER TABLE "physician" ADD FOREIGN KEY ("departmentname") REFERENCES "department" ("departmentname") ON DELETE CASCADE ON UPDATE CASCADE;
