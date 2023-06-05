#!/bin/bash

OUTFILE="get.go"

cat <<EOF > $OUTFILE
// Automatically generated by go generate. DO NOT EDIT
package data

func GetRecipe(item string) *Recipe {
	return d.GetRecipe(item)
}

func GetTech(tech string) *Technology {
	return d.GetTech(tech)
}

EOF

for thing in AssemblingMachine Boiler Furnace Generator Item Lab Module RocketSilo; do
	cat <<EOF >> $OUTFILE
func Get$thing(name string) *$thing {
	x := d.$thing[name]
	return &x
}

EOF
done



