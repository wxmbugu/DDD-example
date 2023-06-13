pwd := ${CURDIR}

postgres:
	docker run --name postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=secret -p 5432:5432 -v ~/postgres_data:/data/db -d postgres:14-alpine
createdb:
	docker exec -it postgres createdb --username=postgres --owner=postgres patient_tracker
startdb:
	docker start postgres
stopdb:
	docker stop postgres
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
migrateforce6:
	docker run -v "$(pwd)/internal/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable"  force 6
migrateforce7:
	docker run -v "$(pwd)/internal/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable"  force 7
migrateforce8:
	docker run -v "$(pwd)/internal/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/patient_tracker?sslmode=disable"  force 8
test:
	go test -v -cover ./...
server:
	go run ./cmd/patient_tracker
admin:
	go run ./cmd/admin
app:
	go build ./cmd/patient_tracker
repl:
	go build ./cmd/admin/

.PHONY: postgres test
