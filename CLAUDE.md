# Project Guidelines

## Project Overview
This is a Golang CLI project that integrates Linear issue tracking with Claude Code.

## Linear API Notes
- Linear uses GraphQL at `https://api.linear.app/graphql`
- Auth header: `Authorization: <API_KEY>` (no Bearer prefix for API keys)
- **Variable types are inconsistent across endpoints:**
  - `issues(filter: { team: { id: { eq: $teamId } } })` expects `$teamId: ID!`
  - `team(id: $teamId)` expects `$teamId: String!`
  - Always check the specific endpoint's expected type

## Working Directory Restriction
Only work within this project folder (`/Users/franzvonderlippe/Programming/linc`). Do not modify files outside this directory.

## External Resources Policy
- **Reading**: You may read external resources (documentation, APIs, etc.)
- **Writing**: Do NOT write to or modify any external resources (databases, APIs, external services, remote repositories, etc.)

## Development Workflow
- **Always rebuild after edits**: Run `go build -o linc` after making code changes
