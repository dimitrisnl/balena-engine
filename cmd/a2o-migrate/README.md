a2o-migrate
===========

migrates images and containers from `aufs` to `overlay2` storage-driver

To enable safe rollback, no breaking changes are applied to the original/future storage locations until we are done. The
overlay2 tree is built in a temporary location: `/var/lib/balena-engine/overlay2.temp` and moved on completion.

We use hardlinks to "duplicate" the layer data. This ensures we have a rollback path at the cost of ~2x the inode
count.

TODO
----

- [ ] migrate the daemon service files
- [ ] implement rollback
