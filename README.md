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

# STOPPING THE SERVER

At the moment there is no control for stopping the server, so you'll
need to kill the process.

# COPYRIGHT AND LICENCE

This project is copyright (c) 2023 Ben Bullock.

It is licenced under the GNU Affero General Public Licence.

