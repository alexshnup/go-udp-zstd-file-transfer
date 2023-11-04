```bash
➜  go-udp-zstd-file-transfer git:(main) ✗ go tool pprof http://localhost:6060/debug/pprof/profile\?seconds\=45
Fetching profile over HTTP from http://localhost:6060/debug/pprof/profile?seconds=45
Saved profile in /Users/alexsh/pprof/pprof.samples.cpu.017.pb.gz
Type: cpu
Time: Nov 4, 2023 at 2:39pm (EDT)
Duration: 45.11s, Total samples = 9.29s (20.59%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 8.51s, 91.60% of 9.29s total
Dropped 163 nodes (cum <= 0.05s)
Showing top 10 nodes out of 88
      flat  flat%   sum%        cum   cum%
     2.59s 27.88% 27.88%      2.59s 27.88%  runtime.pthread_cond_signal
     1.89s 20.34% 48.22%      1.90s 20.45%  syscall.syscall
     1.74s 18.73% 66.95%      1.74s 18.73%  runtime.kevent
     1.24s 13.35% 80.30%      1.24s 13.35%  syscall.syscall6
     0.58s  6.24% 86.54%      0.58s  6.24%  runtime.pthread_cond_wait
     0.12s  1.29% 87.84%      0.12s  1.29%  runtime.pthread_kill
     0.10s  1.08% 88.91%      0.10s  1.08%  runtime.madvise
     0.10s  1.08% 89.99%      0.22s  2.37%  runtime.mallocgc
     0.09s  0.97% 90.96%      0.09s  0.97%  runtime.heapBits.nextFast (inline)
     0.06s  0.65% 91.60%      0.21s  2.26%  runtime.scanobject
(pprof) list runtime.pthread_cond_signal
Total: 9.29s
ROUTINE ======================== runtime.pthread_cond_signal in /usr/local/go/src/runtime/sys_darwin.go
     2.59s      2.59s (flat, cum) 27.88% of Total
         .          .    527:func pthread_cond_signal(c *pthreadcond) int32 {
     2.59s      2.59s    528:   ret := libcCall(unsafe.Pointer(abi.FuncPCABI0(pthread_cond_signal_trampoline)), unsafe.Pointer(&c))
         .          .    529:   KeepAlive(c)
         .          .    530:   return ret
         .          .    531:}
         .          .    532:func pthread_cond_signal_trampoline()
         .          .    533:
(pprof) list syscall.syscall
Total: 9.29s
ROUTINE ======================== syscall.syscall in /usr/local/go/src/runtime/sys_darwin.go
     1.89s      1.90s (flat, cum) 20.45% of Total
         .          .     21:func syscall_syscall(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
         .          .     22:   args := struct{ fn, a1, a2, a3, r1, r2, err uintptr }{fn, a1, a2, a3, r1, r2, err}
         .          .     23:   entersyscall()
     1.89s      1.90s     24:   libcCall(unsafe.Pointer(abi.FuncPCABI0(syscall)), unsafe.Pointer(&args))
         .          .     25:   exitsyscall()
         .          .     26:   return args.r1, args.r2, args.err
         .          .     27:}
         .          .     28:func syscall()
         .          .     29:
ROUTINE ======================== syscall.syscall6 in /usr/local/go/src/runtime/sys_darwin.go
     1.24s      1.24s (flat, cum) 13.35% of Total
         .          .     43:func syscall_syscall6(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr) {
         .          .     44:   args := struct{ fn, a1, a2, a3, a4, a5, a6, r1, r2, err uintptr }{fn, a1, a2, a3, a4, a5, a6, r1, r2, err}
         .          .     45:   entersyscall()
     1.24s      1.24s     46:   libcCall(unsafe.Pointer(abi.FuncPCABI0(syscall6)), unsafe.Pointer(&args))
         .          .     47:   exitsyscall()
         .          .     48:   return args.r1, args.r2, args.err
         .          .     49:}
         .          .     50:func syscall6()
         .          .     51:
(pprof) 
```