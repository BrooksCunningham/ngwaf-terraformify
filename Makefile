clean:
	- rm -r .terraform*
	- rm *.tfstate*
	- rm *.tfstate.backup

run:
	go run .
	- terraform plan -generate-config-out=generated.tf

rerun:
	make clean
	make run