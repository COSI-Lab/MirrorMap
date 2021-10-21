import json

colorlist = []
with open("static/test.txt", 'r') as f:
    for line in f.readlines():
       colorlist.append(line.split(',')[2]) 

print(colorlist)

l = []
for i in range(255):
    l.append(0)

print(l)