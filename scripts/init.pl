#!/usr/bin/env perl
use warnings;
use strict;
use FindBin '$Bin';

use JSON::Parse 'read_json';
use DBI;

my $file = 'bagzulla.db';

chdir "$Bin/.." or die $!;
if (-f $file) {
    rename $file, "$file.backup" or die $!;
}
my $status = system ("sqlite3 -batch $file < schema.txt");
if ($status != 0) {
    die "Failed to create $file due to errors from sqlite3";
}
my $db = DBI->connect("dbi:SQLite:dbname=$file",'','',
		      {RaiseError => 1, AutoCommit => 1});
my $users = read_json ("users.json");
for my $user (@$users) {
    my $p = $user->{pass};
    my $n = $user->{login};
    $db->do ("INSERT INTO person(name,email,password) VALUES('$n','$n\@localhost','$p')");
}
$db->do ("INSERT INTO txt(content,entered) VALUES('Project unspecified',CURRENT_TIMESTAMP)");
$db->do ("INSERT INTO project(name,directory,description,owner,status) VALUES('None','',1,1,0)");
$db->disconnect ();
exit;

