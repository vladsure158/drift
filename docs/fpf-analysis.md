# FPF-анализ: менеджер вайб-код проектов (vibecmd)

- Дата: 2026-03-19
- Методология: FPF (First Principles Framework), B.5 Canonical Reasoning Cycle + B.5.2 Abductive Loop

---

## 1. Фрейминг аномалии (B.5.2 Step 1)

**Проблема:** Вайб-кодер создает несколько проектов в день с помощью Claude Code / AI-ассистентов. Каждый проект — быстрый, часто эфемерный. Но **нет единого места**, где видны:
- все проекты и их статус
- процент готовности
- заметки/контекст
- когда последний раз работал
- что осталось доделать

**Ключевое отличие от классического PM:** вайб-код проекты летучие, их много, они создаются быстро. Jira/Linear/Notion — слишком тяжелые. Нужен инструмент на уровне `lazygit` или `htop`, а не Jira.

---

## 2. Кандидаты-гипотезы (B.5.2 Step 2)

| # | Гипотеза | Описание |
|---|----------|----------|
| H1 | **Slash-команда в Claude Code** | `/projects` внутри Claude — показывает список, статус |
| H2 | **Standalone TUI** (терминальный UI) | Отдельное приложение типа Norton Commander / Midnight Commander для проектов |
| H3 | **Web-дашборд** (local-first) | Локальный веб-сервер с UI для управления проектами |
| H4 | **CLI + файловый формат** | `vc list`, `vc status` — простой CLI, данные в `.vibecmd/` в каждом проекте |
| H5 | **Гибрид: CLI + TUI + опциональный веб** | CLI для автоматизации, TUI для интерактива, веб для красивого обзора |
| H6 | **MCP-сервер для Claude Code** | Claude сам читает/пишет метаданные проектов через MCP |

---

## 3. Фильтры правдоподобия (B.5.2 Step 3)

| Критерий | H1 Slash | H2 TUI | H3 Web | H4 CLI | H5 Гибрид | H6 MCP |
|----------|----------|--------|--------|--------|------------|--------|
| **Простота (Parsimony)** | Ограничен API Claude | Средняя | Сложно | Просто | Средне | Просто |
| **Решает проблему** | Частично (только внутри Claude) | Да | Да | Частично (нет обзора) | Да | Частично |
| **Fit с workflow** | Отлично (уже в терминале) | Отлично | Требует браузер | Хорошо | Отлично | Отлично |
| **Скорость MVP** | Быстро, но ограничен | 2-3 недели | 3-4 недели | 1 неделя | 3-4 недели | 1-2 недели |
| **Open-source потенциал** | Низкий (привязка к Claude) | Высокий | Высокий | Высокий | Максимальный | Средний |
| **Масштабируемость** | Нет | Средняя | Высокая | Высокая | Максимальная | Средняя |

---

## 4. Оценка и выбор (B.5.2 Step 4)

**Отсеиваем:**
- **H1 (slash-команда)** — Claude Code не предназначен для persistent UI. Сессия заканчивается = данные теряют контекст отображения. Можно использовать как *дополнение*, но не как основу.
- **H3 (только веб)** — вайб-кодеры живут в терминале. Переключение в браузер — трение. Веб хорош как *опция*, не как основа.
- **H6 (только MCP)** — хорош как интеграция, но не как самостоятельный продукт.

**Prime Hypothesis: H5 (Гибрид) с ядром H4 (CLI + файловый формат)**

---

## 5. Архитектура продукта (MethodDescription, A.3.2)

**Имя:** `vibecmd` (или `vpm` — vibe project manager, или `nc` — nodding to Norton Commander)

### 5.1 Модель данных (Holon, A.1)

Каждый проект = холон с метаданными:

```
~/.vibecmd/registry.json          # Центральный реестр всех проектов
<project>/.vibecmd/project.json   # Метаданные конкретного проекта
```

**project.json:**
```jsonc
{
  "id": "uuid",
  "name": "landing-saas",
  "description": "Лендинг для AI-стартапа",
  "status": "in-progress",    // idea | in-progress | paused | done | abandoned
  "progress": 65,             // 0-100, можно auto + manual
  "tags": ["next.js", "ai", "landing"],
  "created": "2026-03-19T10:00:00Z",
  "lastActivity": "2026-03-19T14:30:00Z",
  "notes": [
    { "ts": "2026-03-19T12:00:00Z", "text": "Добавил hero секцию" }
  ],
  "goals": [
    { "text": "Hero section", "done": true },
    { "text": "Pricing", "done": false },
    { "text": "Deploy to Vercel", "done": false }
  ],
  "stack": ["next.js", "tailwind", "shadcn"],
  "deployUrl": "https://...",
  "repo": "https://github.com/..."
}
```

### 5.2 Слои (холонная композиция)

```
┌──────────────────────────────────────────┐
│  Layer 4: Web Dashboard (optional)       │  ← Next.js app, красивый обзор
├──────────────────────────────────────────┤
│  Layer 3: TUI (interactive)              │  ← Norton Commander-style двухпанельный UI
├──────────────────────────────────────────┤
│  Layer 2: CLI (scriptable)               │  ← vibecmd add/list/status/note/goal
├──────────────────────────────────────────┤
│  Layer 1: Core Library + File Format     │  ← .vibecmd/ + registry, auto-discovery
├──────────────────────────────────────────┤
│  Layer 0: Integrations                   │  ← git auto-detect, CLAUDE.md parser,
│                                          │     MCP server, Claude Code hooks
└──────────────────────────────────────────┘
```

### 5.3 Auto-discovery (ключевая фича)

Инструмент должен **автоматически** обнаруживать проекты и обогащать данные:

| Источник | Что извлекаем |
|----------|---------------|
| `git log` | Последний коммит, частота активности |
| `CLAUDE.md` | Описание проекта, инструкции |
| `package.json` / `pyproject.toml` | Стек, зависимости, скрипты |
| `.vercel/` | Deploy URL, статус |
| Git status | Есть ли uncommitted changes |
| Маркеры в коде (`TODO`, `FIXME`) | Оставшаяся работа |

### 5.4 TUI — Norton Commander стиль

```
┌─ Projects ──────────────────────┬─ Details ──────────────────────┐
│ ▸ landing-saas     65% ●in-pr  │ Name: landing-saas             │
│   ai-chatbot       90% ●in-pr  │ Stack: next.js, tailwind       │
│   portfolio-v2     100% ✓done  │ Progress: ████████░░ 65%       │
│   crypto-tracker   20% ○idea   │ Last: 2h ago                   │
│   email-tool       0%  ◊paused │                                │
│   api-gateway      45% ●in-pr  │ Goals:                         │
│                                │ ✓ Hero section                 │
│                                │ ○ Pricing                      │
│                                │ ○ Deploy to Vercel             │
│                                │                                │
│                                │ Notes:                         │
│                                │ 12:00 Добавил hero секцию      │
├─ Actions ──────────────────────┴────────────────────────────────┤
│ [a]dd  [n]ote  [g]oal  [s]tatus  [o]pen  [d]elete  [q]uit     │
└─────────────────────────────────────────────────────────────────┘
```

### 5.5 CLI — для скриптов и быстрых действий

```bash
vibecmd init                      # Инициализировать проект в текущей директории
vibecmd add ~/projects/new-app    # Добавить существующий проект
vibecmd list                      # Показать все проекты (table format)
vibecmd list --json               # Для скриптов
vibecmd status                    # Статус текущего проекта
vibecmd note "Добавил auth"       # Быстрая заметка
vibecmd goal "Добавить оплату"    # Добавить цель
vibecmd goal done 2               # Отметить цель выполненной
vibecmd progress 75               # Установить прогресс
vibecmd scan ~/Develop            # Найти все проекты в директории
vibecmd tui                       # Открыть интерактивный TUI
vibecmd serve                     # Запустить веб-дашборд (localhost:3333)
```

### 5.6 Claude Code интеграция (Layer 0)

**Вариант A — MCP Server:**
```bash
claude mcp add vibecmd -- vibecmd mcp-serve
```
Тогда Claude может: "покажи мои проекты", "добавь заметку к landing-saas", "какой прогресс по ai-chatbot?"

**Вариант B — Claude Code Hook (SessionStart):**
```json
{
  "SessionStart": [{
    "matcher": "startup",
    "hooks": [{ "type": "command", "command": "vibecmd context-inject" }]
  }]
}
```
При старте Claude сессии автоматически инжектится контекст текущего проекта.

**Вариант C — CLAUDE.md интеграция:**
Vibecmd автогенерирует секцию в CLAUDE.md с текущим статусом проекта, целями, заметками.

---

## 6. Стратегия реализации (WorkPlan, A.15.2)

**Фаза 1 — MVP (1 неделя):**
- Core library (TypeScript/Node.js)
- Файловый формат `.vibecmd/project.json` + `~/.vibecmd/registry.json`
- CLI: `init`, `list`, `status`, `note`, `goal`, `progress`, `scan`
- Auto-discovery: git, package.json

**Фаза 2 — TUI (неделя 2):**
- Двухпанельный TUI на ink или blessed
- Навигация, фильтры, сортировка
- Quick actions (hotkeys)

**Фаза 3 — Интеграции (неделя 3):**
- MCP-сервер для Claude Code
- Claude Code hooks
- Auto-enrichment (git, Vercel, TODOs)

**Фаза 4 — Web Dashboard (опционально):**
- Next.js app, reads from тех же `.vibecmd/` файлов
- Визуализация прогресса, timeline
- Красивый overview для "portfolio review"

---

## 7. Выводы

**Ответ на вопрос "сабкоманда внутри Claude или нет?":**

Сабкоманда внутри Claude — **дополнение, не основа**. Причины:
1. Claude Code сессия не persistent — нет постоянного UI
2. Терминальный вывод Claude ограничен (нет интерактивности)
3. Привязка к одному AI-инструменту снижает open-source ценность

**Оптимальная архитектура:** standalone CLI + TUI (основа) + MCP/hooks (интеграция с Claude) + web (опция).

**Ключевые дифференциаторы от обычных PM-инструментов:**
1. **Auto-discovery** — сканирует директории, находит проекты сам
2. **AI-aware** — понимает CLAUDE.md, интегрируется через MCP
3. **Zero-friction** — `vibecmd init` и готово, никаких настроек
4. **Terminal-native** — TUI вместо браузера, fit вайб-кодеру
5. **File-based** — никаких баз данных, никаких серверов, git-friendly
