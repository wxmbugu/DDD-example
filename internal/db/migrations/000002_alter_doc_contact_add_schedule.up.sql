CREATE TABLE IF NOT EXISTS "schedule" (
  "scheduleid" SERIAL PRIMARY KEY,
  "doctorid" integer NOT NULL,
  "starttime" timestamp NOT NULL,
  "endtime" timestamp NOT NULL
);
CREATE INDEX ON "schedule" ("scheduleid");
ALTER TABLE IF EXISTS "physician" ADD COLUMN "contact" varchar UNIQUE NOT NULL;
ALTER TABLE "schedule" ADD FOREIGN KEY ("doctorid") REFERENCES "physician" ("doctorid");




