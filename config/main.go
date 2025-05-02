package config

var Host string = "0.0.0.0"
var Port int = 7379
var KEYS_LIMIT = 100
var EVICTION_STRATEGY = "EVICT_LRU"
var APPEND_ONLY_FILE = "dice.aof"
var EVICTION_RATIO = 0.4
var SAMPLE_SIZE = 20
var EVICTION_POOL_SIZE = 16
