# Routing

### BEN HOYT, Different approaches to HTTP routing in Go
https://benhoyt.com/writings/go-routing/

### Axel Wagner, How to not use an http-router in go
https://blog.merovius.de/2017/06/18/how-not-to-use-an-http-router.html


### julienschmidt / httprouter
https://github.com/julienschmidt/httprouter

### httptreemux 
https://nicedoc.io/dimfeld/httptreemux
This is inspired by Julien Schmidt's httprouter, in that it uses a patricia tree, but the implementation is rather different. Specifically, the routing rules are relaxed so that a single path segment may be a wildcard in one route and a static token in another. This gives a nice combination of high performance with a lot of convenience in designing the routing patterns. In benchmarks, httptreemux is close to, but slightly slower than, httprouter.