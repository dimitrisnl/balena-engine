a2o-migrate
===========

migrates images and containers from `aufs` to `overlay2` storage-driver

To enable safe rollback, no breaking changes are applied to the original/future storage locations until we are done. The
overlay2 tree is built in a temporary location: `/var/lib/balena-engine/overlay2.temp` and moved on completion.

We use hardlinks to "duplicate" the layer data. This ensures we have a rollback path at the cost of ~2x the inode
count.


usage
-----

To run the migration:
```
$ a2o-migrate -migrate
```

Run some containers, verify everything works as expected and then:
```
$ a2o-migrate -cleanup
```

Or if something went wrong / you want to go back:
```
$ a2o-migrate -rollback
```
