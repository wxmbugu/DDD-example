# patient management system
[![Go](https://github.com/Wambug/patient-management-system/actions/workflows/go.yml/badge.svg)](https://github.com/Wambug/patient-management-system/actions/workflows/go.yml)

# Patient Tracker System(WIP)
  - This is a simple patient management system that tracks appointment and also tracks the patient documents.

### Prerequisite
Ensure you have the following installed:
 1. Golang
 2. Postgres
 3. Make

#### Initial Setup
  - Note to access the db run :- make accessdb
  - If you run to any to any complications when running make migrateup 
   run check the version and fix it accordingly :- make migrateforce($version) e.g make migrateforce1 fixes version one of migrating up or down
   1. Setup Db
``` 
$ git clone https://github.com/wxmbugu/DDD-example.git
$ cd DDD-example
$ make postgres ## installs postgres docker image.
$ make createdb
$ make startdb 
```
2. Setup Migrate 
```
$ make migrate ## installs migrate
$ make migratup ## runs migrations
$ make migratedown 
```
 3. Run  Server
```
$ make server
```
 4. Test 
```
 $ make test
```
5. Run Amin REPL
```
$ make admin
 use the help commmand inside the repl.
```

#### TODO
- [ ] Search Functionality (engine)
- [ ] Verification
- [ ] Set Reminder for appointments
- [x] Logger (wip)
- [ ] Avatar Upload
- [ ] Increase Test Suite Coverage 
