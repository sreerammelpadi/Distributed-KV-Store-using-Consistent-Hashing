import time
import grpc
import helloworld_pb2
import helloworld_pb2_grpc

def run():
    channel = grpc.insecure_channel("localhost:50051")
    stub = helloworld_pb2_grpc.GreeterStub(channel)

    s = time.time()
    for i in range(1000):
        resp = stub.SayHello(helloworld_pb2.HelloRequest(name=f"Call {i}"))
        # print(resp.message)
    print(f"Got all responses in {time.time() - s} seconds")

if __name__ == "__main__":
    run()
