# tpider

&copy;[Mentu](mailto:mentu.zhou@outlook.com)

***Spider for tumblr.***

---

## Installation

It depends nothing, just `go get` it.

```bash
$ go get github.com/TinkerBravo/tpider
```

## Getting started

Download picture and video of tumblr user "staff" to current path, with 3 concurrent threads:

```console
$ tpider -user="staff" -proxy="" -path="." -thread=3
```

If you are unluckly behind the great wall, please get a proxy.