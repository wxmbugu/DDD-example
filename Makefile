postgres:
	docker run --name postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=secret -p 5432:5432 -v ~/postgres_data:/data/db -d postgres:14-alpine
createdb:
	docker exec -it postgres createdb --username=postgres --owner=postgres patient_tracker
startdb:
	docker start postgres
accessdb:
	docker exec -it postgres psql -U postgres patient_tracker
dropdb:
	docker exec -it postgres dropdb patient_tracker
migrate:
	docker pull migrate/migrate
migrateup:
	 migrate -path pkg/db/migrations -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable" -verbose up
migratedown:
	migrate -path pkg/db/migrations -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable" -verbose down

.PHONY: postgres startdb accessdb dropdb migrate migrateup migratedown
