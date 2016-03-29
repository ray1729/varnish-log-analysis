#!/usr/bin/env perl

use strict;
use warnings FATAL => 'all';

use Date::Parse qw(str2time);
use List::Util qw(max min);

my $log_line_rx = qr{^[^|]+\|-\|[^|]+\|\[([^]]+)\]\|.+\|(\d+)\|([^|]+)\s*$};

my $wanted_backend_rx = qr/^live_wanda_\d+_cantor\s*$/;

my $window_duration = 5*60;

sub window {
    my $time = shift;
    return int($time/$window_duration)*$window_duration;
}

sub parse_line {
    chomp(my $line = shift);
    #my ($date_str, $duration, $backend) = (split /\|/, $line)[3,-2,-1];
    #for ($date_str) { s/^\[//; s/\]$//; }
    my ($date_str, $duration, $backend) = $line =~ $log_line_rx
        or return;
    if ($backend =~ $wanted_backend_rx) {
        return {date => str2time($date_str),
                duration => $duration,
                backend  => $backend};

    }
    return;
}

my %accum;

while (<>) {
    my $entry = parse_line($_) or next;
    my $window = window($entry->{date});
    if ( $accum{$window} ) {
        $accum{$window}{num_requests}++;
        $accum{$window}{total_duration} += $entry->{duration};
        $accum{$window}{max_duration} = max($accum{$window}{max_duration}, $entry->{duration});
        $accum{$window}{min_duration} = min($accum{$window}{min_duration}, $entry->{duration});

    }
    else {
        $accum{$window}{num_requests} = 1;
        for ( qw(min_duration max_duration total_duration) ) {
            $accum{$window}{$_} = $entry->{duration};

        }
    }
}

for my $k ( sort keys %accum ) {
    my $entry = $accum{$k};
    printf( "%s % 8d % 10d % 10d % 10d\n",
            scalar(localtime $k),
            $entry->{num_requests},
            $entry->{min_duration},
            $entry->{max_duration},
            int($entry->{total_duration}/$entry->{num_requests})
        );
}
