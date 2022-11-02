# udsplay
playground

# results

- Unix sockets are removed when the creating process exits *cleanly*
- Unix sockets can be symlinked
  - When the original socket is removed and recreated, the symlink works again
- Unix sockets can be bind mounted
  - When the original socket is removed and recreated, the bind mount is not updated


- readers get an EOF when the listener exits cleanly or uncleanly
