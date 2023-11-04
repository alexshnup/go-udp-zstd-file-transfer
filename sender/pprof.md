```bash
➜  go-udp-zstd-file-transfer git:(main) ✗ go tool pprof http://localhost:6061/debug/pprof/profile\?seconds\=45
Fetching profile over HTTP from http://localhost:6061/debug/pprof/profile?seconds=45
Saved profile in /Users/alexsh/pprof/pprof.samples.cpu.016.pb.gz
Type: cpu
Time: Nov 4, 2023 at 2:37pm (EDT)
Duration: 45.12s, Total samples = 56.49s (125.20%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 47.06s, 83.31% of 56.49s total
Dropped 229 nodes (cum <= 0.28s)
Showing top 10 nodes out of 115
      flat  flat%   sum%        cum   cum%
     9.33s 16.52% 16.52%      9.33s 16.52%  runtime.madvise
     8.53s 15.10% 31.62%      8.53s 15.10%  runtime.kevent
     6.50s 11.51% 43.12%      6.50s 11.51%  runtime.pthread_cond_wait
     5.82s 10.30% 53.43%      5.82s 10.30%  runtime.pthread_kill
     5.18s  9.17% 62.60%      5.19s  9.19%  syscall.syscall6
     3.82s  6.76% 69.36%      3.82s  6.76%  runtime.pthread_cond_signal
     2.36s  4.18% 73.54%      6.65s 11.77%  runtime.scanobject
     1.96s  3.47% 77.00%      1.97s  3.49%  runtime.usleep
     1.91s  3.38% 80.39%      1.91s  3.38%  runtime.heapBits.nextFast (inline)
     1.65s  2.92% 83.31%      1.65s  2.92%  runtime.pthread_cond_timedwait_relative_np
(pprof) list runtime.madvise
Total: 56.49s
ROUTINE ======================== runtime.madvise in /usr/local/go/src/runtime/sys_darwin.go
     9.33s      9.33s (flat, cum) 16.52% of Total
         .          .    252:func madvise(addr unsafe.Pointer, n uintptr, flags int32) {
     9.33s      9.33s    253:   libcCall(unsafe.Pointer(abi.FuncPCABI0(madvise_trampoline)), unsafe.Pointer(&addr))
         .          .    254:   KeepAlive(addr) // Just for consistency. Hopefully addr is not a Go address.
         .          .    255:}
         .          .    256:func madvise_trampoline()
         .          .    257:
         .          .    258://go:nosplit
(pprof) 
```