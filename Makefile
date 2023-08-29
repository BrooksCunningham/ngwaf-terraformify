clean:
	- rm -r .terraform*
	- rm *.tf
	- rm *.tfstate*
	- rm *.tfstate.backup

run:
	go run .

rerun:
	make clean
	make run