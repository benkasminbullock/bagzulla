package Bagzulla;
use parent Exporter;
our @EXPORT_OK = qw/read_statuses connect/;
our %EXPORT_TAGS = (all => \@EXPORT_OK);
use warnings;
use strict;
use utf8;
use FindBin '$Bin';
use Carp;
use File::Slurper 'read_lines';
use DBI;

# $dir is the directory this module is in.

my $dir = __FILE__;
$dir =~ s!/[^/]*$!!;
#die "No '$dir'" unless -d $dir;

=head2 read_statuses

    my @statuses = read_statuses ();

Read the data file of statuses

=cut

sub read_statuses
{
    my $file = "$dir/../statuses.txt";
    if (! -f $file) {
	croak "No $file";
    }
    my @lines = read_lines ($file);
    @lines = map {s/\s//gr} @lines;
    @lines = grep /\S/, @lines;
    return @lines;
}


=head2 connect

    my $dbh = connect ();

Return a database handler which is connected to the SQLite database of
bagzulla. This connects using Perl's Unicode encoding.

=cut

sub connect
{
    my $dbh = DBI->connect (
	"dbi:SQLite:$dir/bagzulla.db",
	undef, undef,
	{
	    RaiseError => 1,
	    sqlite_unicode => 1,
	});
    if (! $dbh) {
	croak DBI->error ();
    }
    return $dbh;
}

1;
