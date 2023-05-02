ALTER TABLE IF EXISTS "patientrecords" ADD COLUMN "nurseid" int NOT NULL; 
ALTER TABLE "patientrecords" ADD FOREIGN KEY ("nurseid") REFERENCES "nurse" ("id") ON DELETE CASCADE;

