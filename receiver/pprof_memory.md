# Memory Profiling

```bash
➜  go-udp-zstd-file-transfer git:(main) ✗ go tool pprof http://localhost:6060/debug/pprof/heap\?seconds\=45
Fetching profile over HTTP from http://localhost:6060/debug/pprof/heap?seconds=45
Saved profile in /Users/alexsh/pprof/pprof.alloc_objects.alloc_space.inuse_objects.inuse_space.002.pb.gz
Type: inuse_space
Time: Nov 4, 2023 at 2:53pm (EDT)
Duration: 45s, Total samples = 512.02kB 
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for -512.02kB, 100% of 512.02kB total
Showing top 10 nodes out of 12
      flat  flat%   sum%        cum   cum%
 -512.02kB   100%   100%  -512.02kB   100%  encoding/gob.(*Decoder).compileDec
         0     0%   100%  -512.02kB   100%  encoding/gob.(*Decoder).Decode
         0     0%   100%  -512.02kB   100%  encoding/gob.(*Decoder).DecodeValue
         0     0%   100%  -512.02kB   100%  encoding/gob.(*Decoder).decOpFor
         0     0%   100%  -512.02kB   100%  encoding/gob.(*Decoder).decodeTypeSequence
         0     0%   100%  -512.02kB   100%  encoding/gob.(*Decoder).decodeValue
         0     0%   100%  -512.02kB   100%  encoding/gob.(*Decoder).getDecEnginePtr
         0     0%   100%  -512.02kB   100%  encoding/gob.(*Decoder).recvType
         0     0%   100%  -512.02kB   100%  main.deserialize
         0     0%   100%  -512.02kB   100%  main.main
(pprof)
```


