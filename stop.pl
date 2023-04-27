#!/home/ben/software/install/bin/perl
use warnings;
use strict;
use utf8;
use FindBin '$Bin';
use LWP::UserAgent;
my $ua = LWP::UserAgent->new ();
$ua->get ("http://mikan/bagpub/controls/?stop=1");
#use BKB::Stuff 'kill_match';
#kill_match (qr!/bagzulla.*bagpub!);
