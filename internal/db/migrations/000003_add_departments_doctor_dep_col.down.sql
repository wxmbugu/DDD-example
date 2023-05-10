DROP TABLE IF EXISTS department CASCADE;
ALTER TABLE "physician" DROP CONSTRAINT IF EXISTS departmentname;
ALTER TABLE "physician" DROP COLUMN IF EXISTS "departmentname";
