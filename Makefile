postgres:
	docker run --name patient-tracker -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=secret -p 5432:5432 -v ~/postgres_data:/data/db -d postgres:14-alpine
startdb:
	docker exec -it patient-tracker /bin/sh
.PHONY: postgres startdb