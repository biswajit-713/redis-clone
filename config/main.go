package config

var Host string = "0.0.0.0"
var Port int = 7379
var KeysLimit = 100
var EvictionStrategy = "EVICT_RANDOM"
var AppendOnlyFile = "dice.aof"
var EvictionRatio = 0.4
