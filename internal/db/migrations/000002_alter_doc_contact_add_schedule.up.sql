DROP TYPE IF EXISTS scheduletype;
CREATE TABLE IF NOT EXISTS "schedule" (
  "scheduleid" SERIAL PRIMARY KEY,
  "doctorid" integer NOT NULL,
  "starttime" varchar NOT NULL,
  "endtime" varchar NOT NULL,
  "active" boolean 
);
CREATE INDEX ON "schedule" ("scheduleid");
ALTER TABLE IF EXISTS "physician" ADD COLUMN "contact" varchar UNIQUE NOT NULL;
ALTER TABLE "schedule" ADD FOREIGN KEY ("doctorid") REFERENCES "physician" ("doctorid") ON DELETE CASCADE ;




