# Fastly's Next-Gen WAF Terraform Configuration generation tool

This tool imports your NGWAF settings to the Terraform state and generates the Terraform configuration for your NGWAF config in HCL.

## Feature list and status
- [x] Corp Rules                :white_check_mark:
- [x] Corp Lists                :white_check_mark:
- [x] Corp Signals              :white_check_mark:
- [x] Site Request Rules        :white_check_mark:
- [x] Site Rate Limiting Rules  :white_check_mark:
- [x] Site Templated Rules      :white_check_mark:
- [x] Site Lists                :white_check_mark:
- [x] Site Signals              :white_check_mark:
- [x] Site Alerts               :white_check_mark:
- [x] Site Agent Alerts         :white_check_mark:
- [ ] Space lasers
- [ ] Coffee maker

# Set up
Environment variables must exist
```
TF_VAR_NGWAF_CORP
TF_VAR_NGWAF_EMAIL
TF_VAR_NGWAF_TOKEN
TF_VAR_SIGSCI_CORP
TF_VAR_SIGSCI_EMAIL
TF_VAR_SIGSCI_TOKEN
```

Just run `make run`


# Need to start over?
`make rerun`

