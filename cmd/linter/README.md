## Custom static analyzer

В проекте реализован собственный статический анализатор.

Проверяет:
- использование panic
- вызов log.Fatal вне main
- вызов os.Exit вне main

Запуск:

go run ./cmd/linter ./...