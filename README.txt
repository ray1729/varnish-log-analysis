%h|%l|%u|%t|%r|%s|%b|%{Referer}i|%{User-agent}i|%{Varnish:handling}x|%D|%{X-Backend}o

%h - remote host
%l - remote logname (always -)
%u - remote user (from auth)
%t - time the request was received
%r - first line of the request
%s - status sent to the client
%b - size of the response, excluding header, in bytes
%{Referer}i - Referer request header
%{User-agent}i - User-agent request header
%{Varnish:handling}x - Cache hit/miss/pass/pipe or error
%D - time taken to serve the request in microseconds
%{X-Backend}o - X-Backend response header

Before optimization, a lot of unnecessary date and duration parsing:

$ time ./main ../varnish.log
real 0m42.392s
user 0m42.314s
sys 0m0.143s

Testing backend before parsing the time and duration:

real 0m43.225s
user 0m42.963s
sys 0m0.312s
