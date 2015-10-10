mkdir test/

thrift \
  --gen go:package_prefix=github.com/upfluence/thrift-http-go/,thrift_import=github.com/upfluence/thrift/lib/go/thrift \
  --out . test.thrift

pushd server

go build
./server &
popd

sleep 2

pushd client
go build
./client
res=$?
popd

kill -9 `cat /tmp/test-http-thrift-go`
echo $res
exit $res
