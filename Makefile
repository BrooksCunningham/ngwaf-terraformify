clean:
	- rm -r .terraform*
	- rm *.tfstate*
	- rm *.tfstate.backup
	- rm generated.tf
	- rm import.tf

run:
	go run .
	- terraform init
	- terraform plan -generate-config-out=generated.tf

rerun:
	make clean
	make run