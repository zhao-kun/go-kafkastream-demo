#!/bin/bash


curl -s http://127.0.0.1:9527/api/v1/labels|jq -r '.[] | [.id, .name, .last_modifier, .creator] | @csv' |\
	awk -v \
	FS="," \
	'BEGIN{print "ID\tName\t\tModifier\t\tCreator";print "======================================================="}{printf "%s\t%s\t\t%s\t\t\t%s\t%s",$1,$2,$3,$4,ORS}'
