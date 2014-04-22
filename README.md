# dockviz

Visualizing Docker Data

This command takes the raw Docker JSON and visualizes it in various ways.  The
only thing that works so far is [Graphviz](http://www.graphviz.org), but others
will come.

# Running

Currently, this only works when the remote API is listening on TCP.  Soon, the Docker command line will allow dumping the image JSON.

```
$ curl -s http://localhost:4243/images/json?all=1 | ./dockviz images -d > images.dot
```

# Download

For now, download binaries from Gobuild: <http://gobuild.io/download/github.com/justone/dockviz>
