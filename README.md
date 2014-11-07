graphstk
========

Generate graphviz from golang stack trace

example usage
=============
```
graphstk > stack.png <<EOF
/home/dgnorton/go/src/github.com/influxdb/influxdb/datastore/shard_datastore.go:189 (0x7c4f90)
(*ShardDatastore).GetOrCreateShard: hopwatch.Printf("err=%s", err.Error()).Break()
/home/dgnorton/go/src/github.com/influxdb/influxdb/cluster/shard.go:154 (0x7187ca)
(*ShardData).SetLocalStore: _, err := self.store.GetOrCreateShard(self.id)
/home/dgnorton/go/src/github.com/influxdb/influxdb/cluster/cluster_configuration.go:1073 (0x71294b)
(*ClusterConfiguration).AddShards: err := shard.SetLocalStore(self.shardStore, self.LocalServer.Id)
/home/dgnorton/go/src/github.com/influxdb/influxdb/coordinator/command.go:359 (0x5a8145)
(*CreateShardsCommand).Apply: createdShards, err := config.AddShards(c.Shards)
EOF
```
