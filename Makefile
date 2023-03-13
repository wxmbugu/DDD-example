pwd := ${CURDIR}

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
	docker run -v "$(pwd)/internal/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable" -verbose up
migratedown:
	docker run -v "$(pwd)/internal/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable" -verbose down -all
migrateforce1:
	docker run -v "$(pwd)/internal/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable" force 1
migrateforce2:
	docker run -v "$(pwd)/internal/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable" force 2
migrateforce3:
	docker run -v "$(pwd)/internal/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable"  force 3
migrateforce4:
	docker run -v "$(pwd)/internal/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable"  force 4
migrateforce5:
	docker run -v "$(pwd)/internal/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable"  force 5
test:
	go test -v -cover ./...
server:
	go run ./cmd/patient_tracker
.PHONY: postgres test
