import json

colorlist = []
with open("static/test.txt", 'r') as f:
    for line in f.readlines():
       colorlist.append(line.split(',')[2]) 

print(colorlist)