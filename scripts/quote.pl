#!/home/ben/software/install/bin/perl
use Z;
my @files = <*.html>;
for my $file (@files) {
    my $text = read_text ($file);
    my $orig = $text;
    $text =~ s!'([^']*?)'!"$1"!gs;
    if ($text ne $orig) {
	print "$file\n"; # --\n$orig\n--\n$text\n--\n";
	write_text ($file, $text);
    }
}
