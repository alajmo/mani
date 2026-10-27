# Filtering Projects

Projects can be filtered when managing projects (sync, list, describe) or running tasks. Filters can be specified through CLI flags or target configurations. The filtering is inclusive, meaning projects must satisfy all specified filters to be included in the results.

Available options:

- **cwd**: include only the project under the current working directory, ignoring all other filters
- **all**: include all projects
- **projects**: Filter by project names
- **paths**: Filter by project paths
- **tags**: Filter by project tags
- **tags_expr**: Filter using tag logic expressions
- **target**: Filter using target

For `mani sync/list/describe`:

- No filters: Targets all projects
- Multiple filters: Select intersection of `projects/paths/tags/tags_expr/target` filters

For `mani run/exec` the precedence is:

1. Runtime flags (highest priority)
2. Target flag configuration (`--target`)
3. Task's default target data (lowest priority)

The default target is named `default` and can be overridden by defining a target named `default` in the config. This only applies for sub-commands `run` and `exec`.

## Tags Expression

Tag expressions allow filtering projects using boolean operations on their tags.
The expression is evaluated for each project's tags to determine if the project should be included.

Operators (in precedence order):

- (): Parentheses for grouping
- !: NOT operator (logical negation)
- &&: AND operator (logical conjunction)
- ||: OR operator (logical disjunction)

Tags in expressions can contain any characters except:

- Whitespace (spaces, tabs, newlines)
- Reserved characters: `(`, `)`, `!`, `&`, `|`

This means tags can include letters, numbers, hyphens, underscores, dots, and other special characters like `@`, `#`, `$`, etc. For example: `my-tag`, `v1.0`, `frontend_v2`, `@scope/package`.

### Example

For example, the expression:

- (main && (dev || prod)) && !test

requires the projects to pass these conditions:

- Must have "main" tag
- Must have either "dev" OR "prod" tag
- Must NOT have "test" tag
