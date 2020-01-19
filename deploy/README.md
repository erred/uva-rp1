## integration testing

### Addressing

`172._2a.__b._cd`:

- `a`: testing set
  - `a = 1` for `test1`
- `b`: load balancing group,
  - `b = 1` for `primary1` and all directly connected secondaries, producer, consumer
- `c`: type of server:
  - `c = 0`: primary (load balancer)
  - `c = 1`: secondary (cache)
  - `c = 2`: producer (server)
  - `c = 3`: consumer (client)
- `d`: server instance, `a = 1, b = 2, ...`
  - `d = 1` for `secondary1a`
  - `d = 2` for `secondary1b`

### test1

```
       watcher1
          │
       primary1
          │
    ┌─────┴─────┐
producer1a   client1a
```

### test2

```
                watcher1
                   │
secondary1a ─┬─ primary1
secondary1b ─┘     │
             ┌─────┴─────┐
         producer1a   client1a
```

### test3

```
                           watcher1
                              │
                   ┌──────────┴───────────┐
secondary1a ─┬─ primary1               primary2 ─┬─ secondary2a
secondary1b ─┘     │                      │      └─ secondary2c
             ┌─────┴─────┐          ┌─────┴─────┐
         producer1a   client1a  producer2a   client2a
```

### test4

```
                           watcher1 ─────────────────────────────────────────────────────── watcher2
                              │                                                                │
                   ┌──────────┴───────────┐                                         ┌──────────┴───────────┐
secondary1a ─┬─ primary1               primary2 ─┬─ secondary2a  secondary3a ─┬─ primary3               primary4 ─┬─ secondary4a
secondary1b ─┘     │                      │      └─ secondary2b  secondary3b ─┘     │                      │      └─ secondary4b
             ┌─────┴─────┐          ┌─────┴─────┐                             ┌─────┴─────┐          ┌─────┴─────┐
         producer1a   client1a  producer2a   client2a                     producer3a   client3a  producer4a   client4a
```
