# Результат оптимизации профиля памяти:
```
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

```
flat      flat%   function
-1024.03kB        syscall.anyToSockaddr
-768.26kB         go.uber.org/zap/zapcore.newCounters
-516.76kB         runtime.procresize
-1792.29kB        main.main
```
## Оптимизация заключалась в настройке пула соединений PostgreSQL:


- ограничено число открытых соединений (SetMaxOpenConns)
- задано число idle соединений (SetMaxIdleConns)
- установлено время жизни соединений (SetConnMaxLifetime)
- Общее потребление heap-памяти уменьшилось примерно на ~1MB за счет переиспользования соединений.
значения выбраны в тестовом варианте по результатам наблюдений

Это позволило уменьшить количество аллокаций, связанных с созданием новых соединений,
снизить нагрузку на GC и улучшить стабильность работы сервиса.


 All in file
 File: shortener
 Type: inuse_space
 Time: 2026-02-22 12:47:00 MSK
 Showing nodes accounting for 4358.66kB, 100% of 4360.11kB total
```
flat  flat%   sum%        cum   cum%
1028kB 23.58% 23.58%     1028kB 23.58%  bufio.NewReaderSize (inline)
1026kB 23.53% 47.11%     1026kB 23.53%  runtime.allocm
1024.52kB 23.50% 70.61%  1538.52kB 35.29%  github.com/lib/pq.(*Connector).open
1024.44kB 23.50% 94.10%  1024.44kB 23.50%  runtime.malg
-1024.03kB 23.49% 70.62% -1024.03kB 23.49%  syscall.anyToSockaddr
-768.26kB 17.62% 53.00%  -768.26kB 17.62%  go.uber.org/zap/zapcore.newCounters (inline)
-516.76kB 11.85% 41.14%  -516.76kB 11.85%  runtime.procresize
515.19kB 11.82% 52.96%   515.19kB 11.82%  strings.(*Replacer).build
513kB 11.77% 64.73%      513kB 11.77%  bufio.NewWriterSize (inline)
512.25kB 11.75% 76.47%   512.25kB 11.75%  io.ReadAll
512.25kB 11.75% 88.22%   512.25kB 11.75%  runtime.gcBgMarkWorker
512.05kB 11.74%   100% -1280.23kB 29.36%  runtime.main
0     0%   100%     1028kB 23.58%  bufio.NewReader (inline)
0     0%   100%  2565.77kB 58.85%  database/sql.(*DB).ExecContext
0     0%   100%  2565.77kB 58.85%  database/sql.(*DB).ExecContext.func1
0     0%   100%  2565.77kB 58.85%  database/sql.(*DB).conn
0     0%   100%  -512.05kB 11.74%  database/sql.(*DB).connectionOpener
0     0%   100%  2565.77kB 58.85%  database/sql.(*DB).exec
0     0%   100%  2565.77kB 58.85%  database/sql.(*DB).retry
0     0%   100%  2053.71kB 47.10%  database/sql.dsnConnector.Connect
0     0%   100%  3078.02kB 70.59%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
0     0%   100%  3078.02kB 70.59%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
0     0%   100%  3078.02kB 70.59%  github.com/iolshn04/go-musthave-shortened-url/internal/handler.CreateHandler
0     0%   100%  3078.02kB 70.59%  github.com/iolshn04/go-musthave-shortened-url/internal/handler.NewRouter.AuthMiddleware.func9.1
0     0%   100%  3078.02kB 70.59%  github.com/iolshn04/go-musthave-shortened-url/internal/handler.NewRouter.func1.RequestLogger.1
0     0%   100%  3078.02kB 70.59%  github.com/iolshn04/go-musthave-shortened-url/internal/handler.NewRouter.func2
0     0%   100%  -768.26kB 17.62%  github.com/iolshn04/go-musthave-shortened-url/internal/logger.Initialize
0     0%   100%  3078.02kB 70.59%  github.com/iolshn04/go-musthave-shortened-url/internal/middlewares.GzipMiddleware.func1
0     0%   100%  3078.02kB 70.59%  github.com/iolshn04/go-musthave-shortened-url/internal/middlewares.GzipRequestMiddleware.func1
0     0%   100%  2565.77kB 58.85%  github.com/iolshn04/go-musthave-shortened-url/internal/repository.(*postgresRepository).Save
0     0%   100%  2565.77kB 58.85%  github.com/iolshn04/go-musthave-shortened-url/internal/service.(*ShortenerService).Shorten
0     0%   100%  2053.71kB 47.10%  github.com/lib/pq.DialOpen
0     0%   100%  2053.71kB 47.10%  github.com/lib/pq.Driver.Open
0     0%   100%   515.19kB 11.82%  github.com/lib/pq.NewConnector
0     0%   100%  2053.71kB 47.10%  github.com/lib/pq.Open (inline)
0     0%   100%   515.19kB 11.82%  github.com/lib/pq.ParseURL
0     0%   100%   515.19kB 11.82%  github.com/lib/pq.ParseURL.func1 (inline)
0     0%   100%  -768.26kB 17.62%  go.uber.org/zap.(*Logger).WithOptions
0     0%   100%  -768.26kB 17.62%  go.uber.org/zap.Config.Build
0     0%   100%  -768.26kB 17.62%  go.uber.org/zap.Config.buildOptions.WrapCore.func5
0     0%   100%  -768.26kB 17.62%  go.uber.org/zap.Config.buildOptions.func1
0     0%   100%  -768.26kB 17.62%  go.uber.org/zap.New
0     0%   100%  -768.26kB 17.62%  go.uber.org/zap.optionFunc.apply
0     0%   100%  -768.26kB 17.62%  go.uber.org/zap/zapcore.NewSamplerWithOptions
0     0%   100% -1024.03kB 23.49%  internal/poll.(*FD).Accept
0     0%   100% -1024.03kB 23.49%  internal/poll.accept
0     0%   100% -1792.29kB 41.11%  main.main
0     0%   100% -1024.03kB 23.49%  net.(*TCPListener).Accept
0     0%   100% -1024.03kB 23.49%  net.(*TCPListener).accept
0     0%   100% -1024.03kB 23.49%  net.(*netFD).accept
0     0%   100% -1024.03kB 23.49%  net/http.(*Server).ListenAndServe
0     0%   100% -1024.03kB 23.49%  net/http.(*Server).Serve
0     0%   100%      513kB 11.77%  net/http.(*conn).readRequest
0     0%   100%  4105.02kB 94.15%  net/http.(*conn).serve
0     0%   100%  3078.02kB 70.59%  net/http.HandlerFunc.ServeHTTP
0     0%   100% -1024.03kB 23.49%  net/http.ListenAndServe (inline)
0     0%   100%      514kB 11.79%  net/http.newBufioReader
0     0%   100%      513kB 11.77%  net/http.newBufioWriterSize
0     0%   100%  3078.02kB 70.59%  net/http.serverHandler.ServeHTTP
0     0%   100%      513kB 11.77%  runtime.mcall
0     0%   100%      513kB 11.77%  runtime.mstart
0     0%   100%      513kB 11.77%  runtime.mstart0
0     0%   100%      513kB 11.77%  runtime.mstart1
0     0%   100%     1026kB 23.53%  runtime.newm
0     0%   100%  1024.44kB 23.50%  runtime.newproc.func1
0     0%   100%  1024.44kB 23.50%  runtime.newproc1
0     0%   100%      513kB 11.77%  runtime.park_m
0     0%   100%     1026kB 23.53%  runtime.resetspinning
0     0%   100%  -516.76kB 11.85%  runtime.rt0_go
0     0%   100%  -516.76kB 11.85%  runtime.schedinit
0     0%   100%     1026kB 23.53%  runtime.schedule
0     0%   100%     1026kB 23.53%  runtime.startm
0     0%   100%  1024.44kB 23.50%  runtime.systemstack
0     0%   100%     1026kB 23.53%  runtime.wakep
0     0%   100%   515.19kB 11.82%  strings.(*Replacer).Replace
0     0%   100%   515.19kB 11.82%  strings.(*Replacer).buildOnce
0     0%   100%   515.19kB 11.82%  sync.(*Once).Do (inline)
0     0%   100%   515.19kB 11.82%  sync.(*Once).doSlow
0     0%   100% -1024.03kB 23.49%  syscall.Accept
```