# import time
# import geoip2.database

# def parser(line):
#     foundIp = line.split(' ', 1)[0]
#     return foundIp

# def fileIn():
#     pList = []
#     i = 0
#     with open('access.log.2', 'r') as f:
#         for line in f.readlines():
#             p = parser(line)
#             pList.append(p)
#             if i % 100000 == 0:
#                 print(i)
#             i+=1
#     return pList

from os import wait
import time

with open('access.log.2') as f:
    for line in f.readlines():
        print(line)
        time.sleep(0.001)
# def main():
#     start = time.time()

#     pList = fileIn()
#     print('Done parsing')
#     longLatList = []
#     i = 0
#     # db = geoip2.Open('GeoLite2-City.mmdb')
#     with geoip2.database.Reader('GeoLite2-City.mmdb') as reader:
#         for ip in pList:
#             result = reader.city(ip)
#             q = (result.location.latitude, result.location.longitude)
#             longLatList.append(q)
#             if i % 100000 == 0:
#                 print(i)
#             i+=1


#     print(time.time() - start)

# if __name__ == "__main__":
#     main()