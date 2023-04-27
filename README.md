# INTRODUCTION

This is a very simple bug tracker. It is designed to be used on a
local server. It provides a very simple replacement for some of the
functions of Bugzilla. It uses SQLite as a backing store for the bugs.

# INSTALLATION

At the moment installation is not very smooth.

You can build the application with the command `make`. You then need
to create and populate a database using the schema with

    sqlite -batch bagzulla.db < schema.txt

and then

    scripts/init.pl

which copies some users from "users.json" in the top directory. You'll
need to add a name and password for whatever user name you want to
use, or you can add those directly to the database.

# STARTING THE SERVER

You can run the server like this:

    ./bagzulla &

If you're using a proxy, you can run it something like this:

    nohup ./bagzulla --url http://localhost/xyz --port 1919 > log 2>&1 &

There is an example script in `run.sh` in the top directory.

If you want to make the directories work, you can specify a directory
display URL using the command-line option --display. See run.sh for an
example of how this works for me locally.

# STOPPING THE SERVER

The server can be stopped from the interface using the control at the
top, or by a command of the form http://localhost/bagzulla?stop=1.

# COPYRIGHT AND LICENCE

This project is copyright (c) 2023 Ben Bullock.

It is licenced under the GNU Affero General Public Licence.

