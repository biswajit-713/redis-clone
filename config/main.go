package config

var Host string = "0.0.0.0"
var Port int = 7379
var KeysLimit = 100
var EvictionStrategy = "evict_first"
var AppendOnlyFile = "dice.aof"
var EvictionRatio = 0.4
