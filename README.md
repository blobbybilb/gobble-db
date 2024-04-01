## gobble-db

Pure go (no cgo) disk-based "struct-oriented" embedded DB, with a **simple**, **friendly**, and **general** API.

Goal: friendliest simple embedded DB for small-to-medium-sized Go projects.

**Simple**: 



## Docs

## Info

### Performance

Performance is not a priority; minimal development overhead is. That said,

### Does it support transactions? Async I/O? ACID?

Nope. Too much complexity for the goal of this project.

### Is it production ready?
If your "production" use case allows you to consider a library as new and not popular as this one,
then yes, it will probably be production-ready enough for your use case. It's meant to be simple enough
that any bugs surface quickly and are easy to fix.