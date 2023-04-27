#!/home/ben/software/install/bin/perl
use warnings;
use strict;
use FindBin '$Bin';
use JSON::Parse 'read_json';
use DBI;
chdir "$Bin/.." or die $!;
unlink 'bagzulla.db' or die $!;
system ("sqlite3 -batch bagzulla.db < schema.txt");
my $db = DBI->connect('dbi:SQLite:dbname=bagzulla.db','','',
		      {RaiseError => 1, AutoCommit => 1});
my $users = read_json ("users.json");
for my $user (@$users) {
my $p = $user->{pass};
my $n = $user->{login};
$db->do ("INSERT INTO person(name,email,password) VALUES('$n','$n\@localhost','$p')");
}
$db->do ("INSERT INTO txt(content) VALUES('Project unspecified')");
$db->do ("INSERT INTO project(name,directory,description,owner,status) VALUES('None','',1,1,0)");
$db->disconnect ();
exit;

