#!/bin/bash

for tag in `git tag -l`; do
	if [[ ${tag} == *"-pre."* ]]; then 
		echo "Deleteing pre tag: ${tag}"
		git push origin :${tag}
		git tag -d ${tag}
	fi
done
