# INTRODUCTION

This is a bug tracker. It is designed to be used on a local server. It
provides a replacement for some of the functions of Bugzilla. It uses
SQLite as a backing store for the bugs.

## Design goals

This bug tracker was designed for me to use on a home web server. It
was designed to be very easy to enter the data without a lot of
restrictions. For example you can enter a bug without any description,
or even without a notification of what project the bug belongs to, and
then fix that up later. I was running bugzilla via the CGI interface
at the time, and got very frustrated with having to enter lots of
extraneous data and also with the big delays of Bugzilla. The name is
a parody of "bugzilla".

# INSTALLATION

At the moment installation is not very smooth.

You can build the application with the command `make`. You then need
to create and populate a database using the schema with

    scripts/init.pl

This requires you to have Perl and the modules `DBI`, `DBD::sqlite`,
and `JSON::Parse`. It renames any old database file called
`bagzulla.db` with the suffix `.backup`, then it creates the database
file again from the schema in `schema.txt` by copying some users from
`users.json` in the top directory. You'll need to add a name and
password for whatever user name you want to use, or you can add those
directly to the database using the `sqlite3` command.

# STARTING THE SERVER

You can run the server like this:

    ./bagzulla &

This will connect to the database in its running directory, so you
need to have read and write permissions for `bagzulla.db` in that
directory. You can specify another database using the option
`--database`.

If you're using a proxy, you can run it something like this:

    nohup ./bagzulla --url http://localhost/xyz --port 1919 > log 2>&1 &

There is an example script in `run.sh` in the top directory.

## Changing the database

You can specify another database using the option
`--database`.

    ./bagzulla --database super.db

This will have to be an sqlite3 database file.

## Displaying directories

If you want to make the directories work, you can specify a directory
display URL using the command-line option --display. See run.sh for an
example of how this works for me locally.

# STOPPING THE SERVER

The server can be stopped from the interface using the control at the
top, or by a command of the form
`http://localhost/bagzulla?stop=1`. There is a simple script `stop.pl`
in the top directory which does this using the Perl module
`LWP::UserAgent`.

# COPYRIGHT AND LICENCE

This project is copyright (c) 2023 Ben Bullock.

It is licenced under the GNU Affero General Public Licence.

