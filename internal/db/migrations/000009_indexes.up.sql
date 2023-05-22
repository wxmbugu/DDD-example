CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS "doctor_name_idx" ON "physician" USING GIN ( "username" gin_trgm_ops);
CREATE INDEX IF NOT EXISTS "doctor_dept_idx" ON "physician" USING GIN ("departmentname"  gin_trgm_ops);
