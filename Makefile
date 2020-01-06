BINARY := "raftdemo"

build:
	rm -rf dist | exit 0
	mkdir -p dist/views
	cd backend && make clean && make
	cp backend/$(BINARY) dist
	cp -r frontend/dist/static dist
	cp frontend/dist/index.html dist/views
	cp frontend/dist/favicon.ico dist/views

clean:
	cd backend && make clean
	rm -rf dist
