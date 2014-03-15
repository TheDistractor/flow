all: setup.json
	go run main.go
	
setup.json: setup.coffee
	coffee setup.coffee >setup.json
