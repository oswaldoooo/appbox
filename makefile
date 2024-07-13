%:cmd/%
	cd $< && go build -o ../../bin/$@ && strip ../../bin/$@