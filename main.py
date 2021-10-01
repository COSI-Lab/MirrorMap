
import time

with open('access.log.2') as f:
    for line in f.readlines():
        print(line)
        time.sleep(0.001)
