#!/home/ben/software/install/bin/perl
use warnings;
use strict;
use utf8;
use FindBin '$Bin';
use lib "$Bin";
use Bagzulla ':all';
use Template;
my @statuses = read_statuses ();
my %vars;
$vars{statuses} = \@statuses;
my $tt = Template->new (
    INCLUDE => [$Bin],
    ABSOLUTE => 1,
    ENCODING => 'utf8',
);
my $file = 'bagzulla-status.go';
my $outfile = "$Bin/../$file";
$tt->process ("$Bin/$file.tmpl", \%vars, $outfile, encoding => 'utf8')
    or die '' . $tt->error ();
exit;
