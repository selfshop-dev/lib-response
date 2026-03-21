# tpl-seed

[![CI](https://github.com/selfshop-dev/tpl-seed/actions/workflows/ci.yml/badge.svg)](https://github.com/selfshop-dev/tpl-seed/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/selfshop-dev/tpl-seed/branch/main/graph/badge.svg)](https://codecov.io/gh/selfshop-dev/tpl-seed)
[![Go Report Card](https://goreportcard.com/badge/github.com/selfshop-dev/tpl-seed)](https://goreportcard.com/report/github.com/selfshop-dev/tpl-seed)
[![Go version](https://img.shields.io/github/go-mod/go-version/selfshop-dev/tpl-seed)](go.mod)
[![License](https://img.shields.io/github/license/selfshop-dev/tpl-seed)](LICENSE)

Базовый шаблонный репозиторий для Go-проектов в организации [selfshop-dev](https://github.com/selfshop-dev).

## Использование шаблона

### 1. Клонировать шаблон

```bash
git clone git@github.com:selfshop-dev/tpl-seed.git tpl-seed
cd tpl-seed
```

### 2. Переименовать модуль

Скрипт `rename.sh` заменяет все вхождения `tpl-seed` во всех файлах и именах директорий:

```bash
chmod +x rename.sh
./rename.sh YOUR_NAME
```

### 3. Применить rulesets

```bash
gh api repos/selfshop-dev/tpl-seed/rulesets \
  --method POST \
  --input <(curl -s https://raw.githubusercontent.com/selfshop-dev/platform/main/rulesets/protect-main-branch.json)
```

### 4. Обновить пример

Удалить placeholder реализацию и тесты:

```bash
rm sum.go sum_test.go calc.go calc_test.go
```

Обновить `doc.go` и `doc_test.go` под контекст нового проекта:

- `doc.go` — стандартное место для package-level комментария. Он отображается на [pkg.go.dev](https://pkg.go.dev) как описание пакета.
- `doc_test.go` — `Example*` функции это живая документация: они запускаются как тесты через `go test` и проверяют что `// Output:` совпадает с реальным выводом.

### 5. Обновить этот README

Описать назначение конкретного проекта.

## Makefile

| Цель | Описание |
|---|---|
| `make code-gen` | Запустить `go generate ./...` |
| `make lint` | Запустить golangci-lint |
| `make test` | Генерация кода + тесты с coverage |
| `make prof` | Собрать профили (cpu, mem, block, mutex) |
| `make prof-view` | Открыть профиль в браузере (`FILE=cpu.out` по умолчанию) |

## Лицензия

[`MIT`](LICENSE) © 2026-present [`selfshop-dev`](https://github.com/selfshop-dev)