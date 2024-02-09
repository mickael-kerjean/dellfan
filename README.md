A program to control the fan of a dell poweredge, it's based on the well known ipmi commands and can be configured like this:

```
//    ^
//    |   manual        automatic
//    |<------------> <-------------
//    |              |
// s2 |              x
//    |             /|
//    |            / |
// s1 |          x/  |
//    |         /|   |
//    |        / |   |
// s0 |------x/  |   |
//    |      |   |   |
//    |      |   |   |
//    -------x---x---x------------>
//           t0  t1  t2

// default:
var (
	t = []float64{
		40, // t0
		60, // t1
		80, // t2
	}
	s = []float64{
		5,  // s0
		15, // s1
		50, // s2
	}
)
```

```
make build
mv dellfan.bin /usr/local/bin/dellfan.bin
mv systemctl /etc/systemd/system/dellfan.service
```
