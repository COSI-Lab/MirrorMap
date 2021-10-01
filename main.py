import time

with open('logs/access.log.2') as f:
    for line in f.readlines():
        print(line)
        time.sleep(0.001)

 
"""
This exists purely as a way to test input from stdin for the go script
"""