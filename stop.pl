#!/home/ben/software/install/bin/perl
use warnings;
use strict;
use LWP::UserAgent;
my $ua = LWP::UserAgent->new ();
$ua->get ("http://mikan/bagpub/controls/?stop=1");
exit;
