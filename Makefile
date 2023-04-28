DBGO=./bagzullaDb/bagzullaDb.go
SRCS= \
auth.go \
bagzulla-status.go \
bagzulla.go \
database.go \
fixstring.go \
user.go \


GOL=/home/ben/projects/gologin

DEPS= \
$(GOL)/login/login.go \
$(GOL)/store/store.go \
$(DBGO) 

bagzulla: $(SRCS) $(DEPS)
	go build -o $@ $(SRCS)

bagzulla-status.go:  scripts/bagzulla-status.go.tmpl scripts/make-statuses.pl scripts/Bagzulla.pm statuses.txt
	perl scripts/make-statuses.pl

test:
	go test

clean:
	rm -f example simple foo.db bagzulla bagzulla-db bagzullaDbtest
	purge -r

