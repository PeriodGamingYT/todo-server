# convinience for me
make:
	clear
	rm -f data.json
	cp temp.json data.json
	go run .

make-no-file:
	clear
	go run .
