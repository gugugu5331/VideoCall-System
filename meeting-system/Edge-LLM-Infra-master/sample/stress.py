import socket
import time
import threading
import argparse
from concurrent.futures import ThreadPoolExecutor
from collections import defaultdict

class ConnectionTester:
    def __init__(self, host, port):
        self.host = host
        self.port = port
        self.results = {
            'total_connections': 0,
            'successful_connections': 0,
            'failed_connections': 0,
            'connection_times': [],
            'errors': defaultdict(int)
        }

    def test_connection(self):
        start_time = time.time()
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(5)  
            sock.connect((self.host, self.port))
            sock.close()
            
            duration = time.time() - start_time
            self.results['successful_connections'] += 1
            self.results['connection_times'].append(duration)
            return True
        except Exception as e:
            error_type = type(e).__name__
            self.results['errors'][error_type] += 1
            self.results['failed_connections'] += 1
            return False
        finally:
            self.results['total_connections'] += 1

def run_test(tester, num_connections, num_threads):
    print(f"Starting connection test to {tester.host}:{tester.port}")
    print(f"Total connections: {num_connections}, Threads: {num_threads}")
    
    start_time = time.time()
    
    with ThreadPoolExecutor(max_workers=num_threads) as executor:
        for _ in range(num_connections):
            executor.submit(tester.test_connection)
    
    total_time = time.time() - start_time
    
    connection_times = tester.results['connection_times']
    avg_time = sum(connection_times) / len(connection_times) if connection_times else 0
    min_time = min(connection_times) if connection_times else 0
    max_time = max(connection_times) if connection_times else 0
    throughput = tester.results['total_connections'] / total_time
    
    print("\nTest Results:")
    print(f"Total time: {total_time:.2f} seconds")
    print(f"Total connections attempted: {tester.results['total_connections']}")
    print(f"Successful connections: {tester.results['successful_connections']}")
    print(f"Failed connections: {tester.results['failed_connections']}")
    print(f"Success rate: {(tester.results['successful_connections']/tester.results['total_connections'])*100:.2f}%")
    print(f"Average connection time: {avg_time*1000:.2f} ms")
    print(f"Min connection time: {min_time*1000:.2f} ms")
    print(f"Max connection time: {max_time*1000:.2f} ms")
    print(f"Throughput: {throughput:.2f} connections/second")
    
    if tester.results['errors']:
        print("\nError Summary:")
        for error, count in tester.results['errors'].items():
            print(f"{error}: {count} occurrences")

def main():
    parser = argparse.ArgumentParser(description='TCP Connection Stress Tester')
    parser.add_argument('--host', type=str, default='localhost', help='Server hostname')
    parser.add_argument('--port', type=int, default=10001, help='Server port')
    parser.add_argument('--connections', type=int, default=100, help='Total connections to attempt')
    parser.add_argument('--threads', type=int, default=10, help='Number of concurrent threads')
    
    args = parser.parse_args()
    
    tester = ConnectionTester(args.host, args.port)
    run_test(tester, args.connections, args.threads)

if __name__ == "__main__":
    main()