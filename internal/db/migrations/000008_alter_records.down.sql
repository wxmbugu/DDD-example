ALTER TABLE IF EXISTS "patientrecords" DROP COLUMN IF EXISTS "nurseid" ;
ALTER TABLE IF EXISTS patientrecords DROP CONSTRAINT IF EXISTS nurseid;
