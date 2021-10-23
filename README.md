# Snake War Server


## protocol


- Packet 

```
|------------------------|
| 6       | 2    | 8     |
| version | type | count |
|------------------------|
````

- Handshake (type=0)

```
|-------------------|
| 16   | 8  | 8     |
| seed | id | count |
|-------------------|
```

- Event (type=1)

```
|--------------------|
| 32   | 8    | 8    |
| tick | type | size |
|--------------------|
```

- Event Start

- Event Pause

- Event Move

```
|----------------|
| 8  | 8         |
| id | direction |
|----------------|
```

- Event Over

```
|----|
| 8  |
| id |
|----|
```
