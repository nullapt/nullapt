# examples/ moved to github.com/nullapt/examples

Skill source code lives in [**github.com/nullapt/examples**](https://github.com/nullapt/examples), one directory per skill. This keeps the platform repo (registry API, CLI, web UI) separate from skill content.

## What's where

| Repo | What it contains |
|---|---|
| [nullapt/nullapt](https://github.com/nullapt/nullapt) | This repo. Registry API, CLI, web UI, SDK glue. |
| [nullapt/examples](https://github.com/nullapt/examples) | Source code for buildable skill examples (Rust, TinyGo, etc.). |
| [nullapt/skills](https://github.com/nullapt/skills) | Curated index of skills published on the registry. |
| [nullapt/sdk](https://github.com/nullapt/sdk) | Language SDKs for skill authors. |

## Building and publishing your own skill

You don't need to clone anything to publish — same as `npm publish`. Work in your own repo, then:

```bash
nullapt keygen                    # one-time keypair generation
nullapt sign ./SKILL.json
nullapt login                     # one-time GitHub OAuth
nullapt publish ./SKILL.json
```

For the skill source layout (Cargo.toml, src/lib.rs, SKILL.json), copy from one of the [example skills](https://github.com/nullapt/examples).
