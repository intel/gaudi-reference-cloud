#!/usr/bin/env python3
#
# Return a dynamic TCP port that is likely to be free.
# Based on https://unix.stackexchange.com/a/132524
#
import socket

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.bind(('', 0))
addr = s.getsockname()
print(addr[1])
s.close()
