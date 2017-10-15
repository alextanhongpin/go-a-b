# go-a/b

A simple a/b metrics data collection and visualization with redis.

## Create a new ab test

```bash
$ curl -XPOST http://localhost:8080/ab/yes-or-no -d '{"options": ["yes", "no"], "name": "Yes or No test", "slug": "yes-or-no", "description": "a simple test to visualize if user picks option yes over no"}'
```

## Incrementing the option

```bash
$ curl -XPOST http://localhost:8080/ab/yes-or-no/yes
```

## Getting the options

```bash
$ curl http://localhost:8080/ab/yes-or-no
```

Output:

```
Yes: 10
No: 0
```

## Getting all ab tests

```bash
$ curl http://localhost:8080/ab
```
