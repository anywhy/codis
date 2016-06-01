#!/usr/bin/python
# -*- encoding:utf-8 -*-
import os
import sys

if sys.version_info >= (2, 6):
    import json

basepath = os.path.abspath(os.path.join(os.path.dirname(__file__), os.pardir))
f = file(os.path.join(basepath, "etc/group.json"))
if sys.version_info >= (2, 6):
    codis_group = json.load(f)
else:
    codis_group = eval(f.read())

nodes = codis_group['codis-group']
# print(nodes)
for item in nodes:
    group_id = item['group_id']
    for node in item['nodes']:
	print ("add redis node:" + node['host'] + " to group:" + str(group_id) )
        #removeCmd = '../bin/codis-config -c ../etc/config.ini server remove-group ' + str(group_id)
        #os.system(removeCmd)
        addCmd = '../bin/codis-config -c ../etc/config.ini server add ' + str(group_id) + ' ' + node['host'] + ' ' + node['type']
        os.system(addCmd)
